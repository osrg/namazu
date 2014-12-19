// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Based on Eli Bendersky's llvm sample code
// https://github.com/eliben/llvm-clang-samples/

#include <string>
#include <fstream>

#include "clang/AST/AST.h"
#include "clang/ASTMatchers/ASTMatchers.h"
#include "clang/ASTMatchers/ASTMatchFinder.h"
#include "clang/Basic/SourceManager.h"
#include "clang/Frontend/TextDiagnosticPrinter.h"
#include "clang/Tooling/CommonOptionsParser.h"
#include "clang/Tooling/Refactoring.h"
#include "clang/Tooling/Tooling.h"
#include "clang/Driver/Options.h"
#include "llvm/Support/raw_ostream.h"

// apt-get install libjsoncpp-dev
#include <jsoncpp/json/json.h>

using namespace clang;
using namespace clang::ast_matchers;
using namespace clang::driver;
using namespace clang::tooling;

using namespace std;

using namespace llvm;

static cl::OptionCategory EqInspectorCategory("earthquake c inspector category");
static cl::opt<bool>
Verbose("verbose", cl::desc("Be verbose"), cl::cat(EqInspectorCategory));
static cl::opt<string>
InspectionListPath("inspection-list-path",
		   cl::desc("Path of inspection list"), cl::cat(EqInspectorCategory));

static vector<string> match_func_calls;

static bool is_matched_func_call(string name) {
  for (auto f: match_func_calls) {
    if (f == name) {
      return true;
    }
  }

  return false;
}

class CallExprHandler : public MatchFinder::MatchCallback {
public:
  CallExprHandler(Replacements *Replace) : Replace(Replace) {}

  virtual void run(const MatchFinder::MatchResult &Result) {
    // The matched 'if' statement was bound to 'ifStmt'.
    if (const CallExpr *call = Result.Nodes.getNodeAs<clang::CallExpr>("callExpr")) {
      const Decl *calleeDecl = call->getCalleeDecl();
      const FunctionDecl *funcDecl = calleeDecl->getAsFunction();
      DeclarationNameInfo nameInfo = funcDecl->getNameInfo();
      string name = nameInfo.getAsString();

      if (!is_matched_func_call(name))
	return;

      string insert_code = "(eq_event_func_call(\"" + name + "\"), ";

      Replacement Head(*(Result.SourceManager), call->getLocStart(), 0, insert_code);
      Replace->insert(Head);

      Replacement Tail(*(Result.SourceManager), call->getLocEnd(), 0, ")");
      Replace->insert(Tail);
    }
  }

private:
  Replacements *Replace;
};

static void InsertHeader(string path) {
  string header =
    "/* below code is inserted by earthquake inspector */\n"
    "#ifndef __EQ_INSPECTION_INSERTED__\n"
    "#define __EQ_INSPECTION_INSERTED__\n"
    "extern void eq_event_func_call(const char *);\n"
    "\n"
    "/* below eq_dep and __eq_nop() are stuff just for making dependency*/\n"
    "extern int eq_dep;\n"
    "static __attribute__((unused)) void __eq_nop(void)\n"
    "{\n"
    "        eq_dep++;\n"
    "}\n"
    "#endif\n"
    "/* inserted code end */\n"
    ;

  ifstream isrcStream;
  isrcStream.open(path);

  isrcStream.seekg (0, isrcStream.end);
  int length = isrcStream.tellg();
  isrcStream.seekg(0, isrcStream.beg);

  char *buf = new char[length];
  isrcStream.read(buf, length);
  isrcStream.close();
  string src = buf;
  delete[] buf;

  ofstream srcStream;
  srcStream.open(path);
  srcStream.seekp(0, ios::beg);

  srcStream << header;
  srcStream << src;
  srcStream.close();
}

static void PrepareInspectionList(string path) {
  ifstream listStream;
  listStream.open(path);

  listStream.seekg(0, listStream.end);
  int length = listStream.tellg();
  listStream.seekg(0, listStream.beg);

  char *buf = new char[length];
  listStream.read(buf, length);
  listStream.close();
  string listString = buf;
  delete[] buf;

  Json::Reader reader;
  Json::Value root;
  if (!reader.parse(listString, root)) {
    outs() << "parsing inspection list file: " << path << "failed\n";
    exit(1);
  }

  for (auto val : root) {
    if (val["type"] == "funcCall") {
      Json::Value param = val["param"];
      Json::Value name = param["name"];
      match_func_calls.push_back(name.asString());
    } else {
      outs() << "unknown event to inspect: " << val["type"].asString() << "\n";
      exit(1);
    }
  }
}

int main(int argc, const char **argv) {
  CommonOptionsParser op(argc, argv, EqInspectorCategory);

  if (InspectionListPath == "") {
    outs() << "specify path of inspection list file\n";
    return 1;
  }

  PrepareInspectionList(InspectionListPath);

  for (auto path : op.getSourcePathList()) {
    InsertHeader(path);
  }

  RefactoringTool Tool(op.getCompilations(), op.getSourcePathList());

  // Set up AST matcher callbacks.
  CallExprHandler HandlerForCallExpr(&Tool.getReplacements());

  MatchFinder Finder;
  Finder.addMatcher(callExpr().bind("callExpr"), &HandlerForCallExpr);

  // Run the tool and collect a list of replacements. We could call runAndSave,
  // which would destructively overwrite the files with their new contents.
  // However, for demonstration purposes it's interesting to show the
  // replacements.
  if (int Result = Tool.runAndSave(newFrontendActionFactory(&Finder).get())) {
    return Result;
  }

  if (Verbose) {
    outs() << "Replacements collected by the tool:\n";
    for (auto &r : Tool.getReplacements()) {
      outs() << r.toString() << "\n";
    }
  }

  return 0;
}

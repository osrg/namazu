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

static cl::OptionCategory EqInspectorCategory("namazu c inspector category");
static cl::opt<bool>
Verbose("verbose", cl::desc("Be verbose"), cl::cat(EqInspectorCategory));
static cl::opt<string>
InspectionListPath("inspection-list-path",
		   cl::desc("Path of inspection list"), cl::cat(EqInspectorCategory));

class funcCall {
public:
  funcCall(string name, bool before, bool after) {
    this->name = name;
    this->before = before;
    this->after = after;
  }

  string name;
  bool before, after;

  string retType;
  std::vector<string> params;

  void setRetType(string);
  void addParam(string);
};

void funcCall::setRetType(string type)
{
  retType = type;
}

void funcCall::addParam(string type)
{
  params.push_back(type);
}

static vector<class funcCall *> match_func_calls;

static class funcCall *is_matched_func_call(string name) {
  for (auto f: match_func_calls) {
    if (!f->name.compare(name)) {
      return f;
    }
  }

  return NULL;
}

class CallExprHandler : public MatchFinder::MatchCallback {
public:
  CallExprHandler(Replacements *Replace) : Replace(Replace) {}

  virtual void run(const MatchFinder::MatchResult &Result) {
    if (const CallExpr *call = Result.Nodes.getNodeAs<clang::CallExpr>("callExpr")) {
      const Decl *calleeDecl = call->getCalleeDecl();
      const FunctionDecl *funcDecl = calleeDecl->getAsFunction();
      DeclarationNameInfo nameInfo = funcDecl->getNameInfo();
      string name = nameInfo.getAsString();

      if (!is_matched_func_call(name))
	return;

      // example, a case of inspecting function func()
      // before: func()
      // after: __eq_wrapped_func()
      string insert_code = "__eq_wrapped_";

      Replacement Head(*(Result.SourceManager), call->getLocStart(), 0, insert_code);
      Replace->insert(Head);
    }
  }

private:
  Replacements *Replace;
};

class FunctionDeclHandler : public MatchFinder::MatchCallback {
public:
  FunctionDeclHandler(Replacements *Replace) : Replace(Replace) {}

  virtual void run(const MatchFinder::MatchResult &Result) {
    if (const FunctionDecl *decl = Result.Nodes.getNodeAs<clang::FunctionDecl>("functionDecl")) {
      DeclarationNameInfo nameInfo = decl->getNameInfo();

      class funcCall *f;
      if (!(f = is_matched_func_call(nameInfo.getName().getAsString())))
	return;

      for (auto parmVar: decl->params()) {
	f->addParam(parmVar->getOriginalType().getAsString());
      }
      f->setRetType(decl->getReturnType().getAsString());
    }
  }

private:
  Replacements *Replace;
};

static void InsertWrappedProtoTypes(ofstream &os)
{
  // FIXME: deathly adhoc...
  // TODO: detecting types of wrapped functions and insert correct header
  os << "#include <sys/types.h>\n";

  for (auto f: match_func_calls) {
    os << f->retType << " ";
    os << "__eq_wrapped_" << f->name << "(";

    int i = 0;
    for (auto p: f->params) {
      if (i) {
	os << ",";
      }
      os << p << " param" << i;
      i++;
    }
    os << ");\n";
  }
}

static void InsertWrappedFunctions(ofstream &os)
{
  for (auto f: match_func_calls) {
    os << f->retType << " ";
    os << "__eq_wrapped_" << f->name << "(";

    int i = 0;
    for (auto p: f->params) {
      if (i) {
	os << ",";
      }
      os << p << " param" << i;
      i++;
    }
    os << ") {\n";

    os << "\teq_event_func_call(\"" << f->name << "\");\n";
    os << "\treturn " << f->name << "(";
    i = 0;
    for (auto p: f->params) {
      if (i) {
	os << ",";
      }
      os << " param" << i;
      i++;
    }
    os << ");\n";
    os << "}\n";
  }
}

static void InsertHeader(string path) {
  string header_double_check =
    "/* below code is inserted by namazu inspector */\n"
    "#ifdef __NMZ_INSPECTION_INSERTED__\n"
    "#error \"more than two inspection\"\n"
    "#else\n"
    "#define __NMZ_INSPECTION_INSERTED__\n"
    "#endif\n"
    "/* inserted code end */\n";

  string header_event_funcs =
    "/* below code is inserted by namazu inspector */\n"
    "extern void eq_event_func_call(const char *);\n"
    "/* inserted code end */\n";

  string footer_make_dep =
    "/* below code is inserted by namazu inspector */\n"
    "\n"
    "/* below eq_dep and __eq_nop() are stuff just for making dependency*/\n"
    "extern int eq_dep;\n"
    "static __attribute__((unused)) void __eq_nop(void)\n"
    "{\n"
    "        eq_dep++;\n"
    "}\n"
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
  string *src = new string(buf, length);
  delete[] buf;

  ofstream srcStream;
  srcStream.open(path);
  srcStream.seekp(0, ios::beg);

  srcStream << header_double_check;
  srcStream << header_event_funcs;
  InsertWrappedProtoTypes(srcStream);
  srcStream << *src;
  InsertWrappedFunctions(srcStream);
  srcStream << footer_make_dep;
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
      if (!param.isMember("name")) {
	outs() << "inspection list file is broken, name field is missing\n";
	exit(1);
      }
      Json::Value name = param["name"];
      if (!name.isString()) {
	outs() << "inspection list file is broken, name field is not typed as string\n";
	exit(1);
      }

      bool before = false;
      if (param.isMember("before")) {
	Json::Value _before = param["before"];
	if (!_before.isBool()) {
	  outs() << "inspection list file is broken, before field is not typed as bool\n";
	  exit(1);
	}
	before = _before.asBool();
      }

      bool after = false;
      if (param.isMember("after")) {
	Json::Value _after = param["after"];
	if (!_after.isBool()) {
	  outs() << "inspection list file is broken, after field is not typed as bool\n";
	  exit(1);
	}
	after = _after.asBool();
      }

      match_func_calls.push_back(new funcCall(name.asString(), before, after));
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

  RefactoringTool Tool(op.getCompilations(), op.getSourcePathList());

  // Set up AST matcher callbacks.
  CallExprHandler HandlerForCallExpr(&Tool.getReplacements());
  FunctionDeclHandler HandlerForFunctionDecl(&Tool.getReplacements());

  MatchFinder Finder;
  Finder.addMatcher(callExpr().bind("callExpr"), &HandlerForCallExpr);
  Finder.addMatcher(functionDecl().bind("functionDecl"), &HandlerForFunctionDecl);

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

  for (auto path : op.getSourcePathList()) {
    InsertHeader(path);
  }

  return 0;
}

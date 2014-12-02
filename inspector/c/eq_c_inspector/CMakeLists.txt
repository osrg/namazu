add_clang_executable(eq_c_inspector
  eq_c_inspector.cpp
  )

set(CMAKE_CXX_FLAGS "-std=c++11")

target_link_libraries(eq_c_inspector clangFrontendTool clangFrontend
  clangDriver clangSerialization clangCodeGen clangParse clangSema
  clangStaticAnalyzerFrontend clangStaticAnalyzerCheckers
  clangStaticAnalyzerCore clangAnalysis clangARCMigrate
   clangRewriteFrontend clangEdit clangAST
  clangASTMatchers clangLex clangBasic clangTooling
  jsoncpp)

# 70        -lclangAST \
#  71        -lclangAnalysis \
#  72        -lclangBasic \
#  73        -lclangDriver \
#  74        -lclangEdit \
#  75        -lclangFrontend \
#  76        -lclangFrontendTool \
#  77        -lclangLex \
#  78        -lclangParse \
#  79        -lclangSema \
#  80        -lclangEdit \
#  81        -lclangASTMatchers \
#  82        -lclangRewrite \
#  83        -lclangRewriteFrontend \
#  84        -lclangStaticAnalyzerFrontend \
#  85        -lclangStaticAnalyzerCheckers \
#  86        -lclangStaticAnalyzerCore \
#  87        -lclangSerialization \
#  88        -lclangTooling \

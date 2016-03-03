#! /bin/bash

git submodule update --init
echo "add_subdirectory(eq_c_inspector)" >> clang.git/CMakeLists.txt
ln -s `pwd`/clang.git llvm.git/tools/clang

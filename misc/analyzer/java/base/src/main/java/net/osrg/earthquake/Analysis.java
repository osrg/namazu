// Copyright (C) 2015 Nippon Telegraph and Telephone Corporation.
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

package net.osrg.namazu;

// FIXME: ExperimentPattern should obsolete Analysis
class Analysis {
    private String klazzName = null;
    private String methodName = null;
    private int line = 0;
    private int covOnSuccess = 0;
    private int covOnFailure = 0;

    Analysis(String klazzName, String methodName, int line) {
        this.klazzName = klazzName;
        this.methodName = methodName;
        this.line = line;
    }

    public String getKlazzName() {
        return klazzName;
    }

    public String getMethodName() {
        return methodName;
    }

    public int getLine() {
        return line;
    }

    public int getCovOnSuccess() {
        return covOnSuccess;
    }

    public void setCovOnSuccess(int covOnSuccess) {
        this.covOnSuccess = covOnSuccess;
    }

    public void addBranchOnSuccess(int branchOnSuccess) {
        this.covOnSuccess += branchOnSuccess;
    }

    public int getCovOnFailure() {
        return covOnFailure;
    }

    public void setCovOnFailure(int covOnFailure) {
        this.covOnFailure = covOnFailure;
    }

    public void addBranchOnFailure(int branchOnFailure) {
        this.covOnFailure += branchOnFailure;
    }

    @Override
    public String toString() {
        return "Analysis{" +
                "klazzName='" + klazzName + '\'' +
                ", methodName='" + methodName + '\'' +
                ", line=" + line +
                ", covOnSuccess=" + covOnSuccess +
                ", covOnFailure=" + covOnFailure +
                '}';
    }
}

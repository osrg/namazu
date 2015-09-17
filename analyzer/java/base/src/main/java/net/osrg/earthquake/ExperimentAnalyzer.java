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

package net.osrg.earthquake;


import com.google.common.collect.Table;
import com.google.common.collect.TreeBasedTable;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.jacoco.core.analysis.*;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.util.Map;
import java.util.TreeMap;

public class ExperimentAnalyzer {
    static final Logger LOG = LogManager.getLogger(ExperimentAnalyzer.class);
    private File storageDir;
    private File classesDir;
    // TODO: use 3D table (klazzName, methodName, lineNumber)
    Table<String, String, Map<Integer, EQAnalysis>> analysisTable;

    public ExperimentAnalyzer(String storagePath, String classesPath) {
        this.storageDir = new File(storagePath);
        this.classesDir = new File(classesPath);
        this.analysisTable = TreeBasedTable.create();
    }

    private static IBundleCoverage getBundleCoverage(Experiment experiment, File classesDir) throws IOException {
        final CoverageBuilder coverageBuilder = new CoverageBuilder();
        final Analyzer analyzer = new Analyzer(
                experiment.getExecFileLoader().getExecutionDataStore(), coverageBuilder);
        analyzer.analyzeAll(classesDir);
        return coverageBuilder.getBundle(experiment.getDirPath());
    }

    public void analyze() throws IOException {
        for (File dir : storageDir.listFiles()) {
            if (!dir.isDirectory()) {
                continue;
            }
            Experiment experiment;
            try {
                experiment = new Experiment(dir.getPath());
            } catch (FileNotFoundException fileNotFoundEx) {
                LOG.warn("Skipping {} ({})", dir, fileNotFoundEx);
                continue;
            }
            LOG.debug("Scanning {}: experiment successful={}",
                    dir, experiment.isSuccessful());
            IBundleCoverage bundle = this.getBundleCoverage(experiment, classesDir);
            this.scanBundleCoverage(experiment, bundle);
        }
    }

    private void scanBundleCoverage(Experiment experiment, IBundleCoverage bundle) throws IOException {
        for (IPackageCoverage pkg : bundle.getPackages()) {
            for (IClassCoverage klazz : pkg.getClasses()) {
                for (IMethodCoverage method : klazz.getMethods()) {
                    this.scanMethodCoverage(experiment, bundle, pkg, klazz, method);
                }
            }
        }
    }

    private void scanMethodCoverage(Experiment experiment,
                                    IBundleCoverage bundle,
                                    IPackageCoverage pkg,
                                    IClassCoverage klazz,
                                    IMethodCoverage method) {
        int first = method.getFirstLine();
        int last = method.getLastLine();
        for (int i = first; i <= last; i++) {
            ILine l = method.getLine(i);
            EQAnalysis analysis = this.prepareAnalysis(klazz.getName(), method.getName(), i);
            int coveredCount = l.getBranchCounter().getCoveredCount();
            if (experiment.isSuccessful()) {
                analysis.addBranchOnSuccess(coveredCount);
            } else {
                analysis.addBranchOnFailure(coveredCount);
            }
        }
    }

    private EQAnalysis prepareAnalysis(String klazzName,
                                       String methodName,
                                       int line) {
        Map<Integer, EQAnalysis> m = this.analysisTable.get(klazzName, methodName);
        if (m == null) {
            m = new TreeMap<>();
            this.analysisTable.put(klazzName, methodName, m);
        }
        EQAnalysis analysis = m.get(line);
        if (analysis == null) {
            analysis = new EQAnalysis(klazzName, methodName, line);
            m.put(line, analysis);
        }
        return analysis;
    }

    public void report() {
        for (Map<Integer, EQAnalysis> mapPerLine : this.analysisTable.values()) {
            boolean hit = false;
            for (EQAnalysis analysis : mapPerLine.values()) {
                // TODO: improve heuristic
                boolean cond = analysis.getBranchOnFailure() > analysis.getBranchOnSuccess() * 10;
                if (cond) {
                    if (!hit) {
                        System.out.printf("Suspicious: %s::%s\n",
                                analysis.getKlazzName().replace("/", "."),
                                analysis.getMethodName());
                        hit = true;
                    }
                    System.out.printf(" - at line %d: branch on success=%d, on failure=%d\n",
                            analysis.getLine(),
                            analysis.getBranchOnSuccess(),
                            analysis.getBranchOnFailure());
                }
            }
        }
    }


    class EQAnalysis {
        private String klazzName = null;
        private String methodName = null;
        private int line = 0;
        private int branchOnSuccess = 0;
        private int branchOnFailure = 0;

        EQAnalysis(String klazzName, String methodName, int line) {
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

        public int getBranchOnSuccess() {
            return branchOnSuccess;
        }

        public void setBranchOnSuccess(int branchOnSuccess) {
            this.branchOnSuccess = branchOnSuccess;
        }

        public void addBranchOnSuccess(int branchOnSuccess) {
            this.branchOnSuccess += branchOnSuccess;
        }

        public int getBranchOnFailure() {
            return branchOnFailure;
        }

        public void setBranchOnFailure(int branchOnFailure) {
            this.branchOnFailure = branchOnFailure;
        }

        public void addBranchOnFailure(int branchOnFailure) {
            this.branchOnFailure += branchOnFailure;
        }
    }
}
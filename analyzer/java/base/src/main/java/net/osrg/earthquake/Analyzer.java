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
import java.util.SortedMap;
import java.util.TreeMap;

public class Analyzer {
    static final Logger LOG = LogManager.getLogger(Analyzer.class);
    private File storageDir;
    private File classesDir;

    // TODO: use 3D table (klazzName, methodName, lineNumber)
    private Table<String, String, SortedMap<Integer, Analysis>> analysisTable;

    public Analyzer(String storagePath, String classesPath) {
        this.storageDir = new File(storagePath);
        this.classesDir = new File(classesPath);
        this.analysisTable = TreeBasedTable.create();
    }

    private static IBundleCoverage getBundleCoverage(Experiment experiment, File classesDir) throws IOException {
        final CoverageBuilder coverageBuilder = new CoverageBuilder();
        final org.jacoco.core.analysis.Analyzer analyzer = new org.jacoco.core.analysis.Analyzer(
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
            Analysis analysis = this.prepareAnalysis(klazz.getName(), method.getName(), i);
            int coveredCount = l.getInstructionCounter().getCoveredCount();
            if (experiment.isSuccessful()) {
                analysis.addBranchOnSuccess(coveredCount);
            } else {
                analysis.addBranchOnFailure(coveredCount);
            }
        }
    }

    private Analysis prepareAnalysis(String klazzName,
                                     String methodName,
                                     int line) {
        SortedMap<Integer, Analysis> m = this.analysisTable.get(klazzName, methodName);
        if (m == null) {
            m = new TreeMap<>();
            this.analysisTable.put(klazzName, methodName, m);
        }
        Analysis analysis = m.get(line);
        if (analysis == null) {
            analysis = new Analysis(klazzName, methodName, line);
            m.put(line, analysis);
        }
        return analysis;
    }

    public Table<String, String, SortedMap<Integer, Analysis>> getAnalysisTable() {
        return analysisTable;
    }

}
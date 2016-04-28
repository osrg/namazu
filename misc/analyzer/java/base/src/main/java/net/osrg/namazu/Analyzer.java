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


import com.google.common.collect.Table;
import com.google.common.collect.TreeBasedTable;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.jacoco.core.analysis.*;
import org.jacoco.core.tools.ExecFileLoader;

import java.io.File;
import java.io.IOException;
import java.util.*;

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

    private static IBundleCoverage[] getBundleCoverages(Experiment experiment, File classesDir) throws IOException {
        List<IBundleCoverage> bundles = new ArrayList<IBundleCoverage>();
        for (ExecFileLoader loader : experiment.getExecFileLoaders()) {
            final CoverageBuilder coverageBuilder = new CoverageBuilder();
            final org.jacoco.core.analysis.Analyzer jacocoAnalyzer =
                    new org.jacoco.core.analysis.Analyzer(
                            loader.getExecutionDataStore(),
                            coverageBuilder);
            jacocoAnalyzer.analyzeAll(classesDir);
            bundles.add(coverageBuilder.getBundle(experiment.getDir().getPath()));
        }
        return bundles.toArray(new IBundleCoverage[0]);
    }

    private File[] listExperimentDirs() {
        File[] dirs = this.storageDir.listFiles();
        Arrays.sort(dirs);
        return dirs;
    }

    public void analyze() throws IOException {
        Set<String> hashSet = new TreeSet<String>();
        int scannedExpCount = 0;
        // TODO: parallelize
        for (File dir : this.listExperimentDirs()) {
            Experiment experiment;
            try {
                experiment = new Experiment(dir);
            } catch (Exception e) {
                LOG.warn("Skipping {} ({})", dir, e);
                continue;
            }
            LOG.debug("Scanning {}: experiment successful={}",
                    dir, experiment.isSuccessful());
            IBundleCoverage[] bundles = this.getBundleCoverages(experiment, classesDir);
            for (int nodeIDAsInt = 0; nodeIDAsInt < bundles.length; nodeIDAsInt++) {
                this.scanBundleCoverage(experiment, bundles[nodeIDAsInt], String.format("%d", nodeIDAsInt));
            }
            hashSet.add(experiment.getPattern().computeHash().toString());
            scannedExpCount++;
            LOG.info("Patterns(at scanning {}): {} {}", dir.getName(), scannedExpCount, hashSet.size());
        }
    }

    private void scanBundleCoverage(Experiment experiment, IBundleCoverage bundle, String nodeID) throws IOException {
        for (IPackageCoverage pkg : bundle.getPackages()) {
            for (IClassCoverage klazz : pkg.getClasses()) {
                for (IMethodCoverage method : klazz.getMethods()) {
                    this.scanMethodCoverage(experiment, bundle, pkg, klazz, method, nodeID);
                }
            }
        }
    }

    private void scanMethodCoverage(Experiment experiment,
                                    IBundleCoverage bundle,
                                    IPackageCoverage pkg,
                                    IClassCoverage klazz,
                                    IMethodCoverage method,
                                    String nodeID) {
        int first = method.getFirstLine();
        int last = method.getLastLine();
        for (int i = first; i <= last; i++) {
            ILine l = method.getLine(i);
            int coveredCount = l.getInstructionCounter().getCoveredCount();
            experiment.getPattern().putCount(klazz.getName(), method.getName(), i,
                    nodeID,
                    experiment.isSuccessful(),
                    coveredCount);

            // FIXME: ExperimentPattern should obsolete Analysis
            Analysis analysis = this.prepareAnalysis(klazz.getName(), method.getName(), i);
            if (experiment.isSuccessful()) {
                analysis.addBranchOnSuccess(coveredCount);
            } else {
                analysis.addBranchOnFailure(coveredCount);
            }
        }
    }

    // FIXME: ExperimentPattern should obsolete Analysis
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

    // FIXME: ExperimentPattern should obsolete Analysis
    public Table<String, String, SortedMap<Integer, Analysis>> getAnalysisTable() {
        return analysisTable;
    }

}
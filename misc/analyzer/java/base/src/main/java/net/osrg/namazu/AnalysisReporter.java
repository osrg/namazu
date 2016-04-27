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

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.util.Collection;
import java.util.Iterator;
import java.util.Map;
import java.util.SortedMap;

public class AnalysisReporter {
    static final Logger LOG = LogManager.getLogger(AnalysisReporter.class);
    private Analyzer analyzer;

    public AnalysisReporter(Analyzer analyzer) {
        this.analyzer = analyzer;
    }

    public void report() {
        final Collection<SortedMap<Integer, Analysis>> maps = this.analyzer.getAnalysisTable().values();
        for (SortedMap<Integer, Analysis> mapPerLine : maps) {
            reportPerMethod(mapPerLine);
        }
    }

    private void reportPerMethod(SortedMap<Integer, Analysis> mapPerLine) {
        int currentRangeBegin = -1;
        int currentRangeLast = -1;
        for (Iterator it = mapPerLine.entrySet().iterator(); it.hasNext(); ) {
            Analysis analysis = ((Map.Entry<Integer, Analysis>) it.next()).getValue();

            // TODO: improve heuristic
            boolean cond = analysis.getCovOnFailure() > analysis.getCovOnSuccess() * 10;
            // LOG.debug("cand={}, analysis={}", cond, analysis);
            if (cond) {
                if (currentRangeBegin < 0) {
                    currentRangeBegin = analysis.getLine();
                }
                currentRangeLast = analysis.getLine();
            }

            if (currentRangeBegin > 0 && (!cond || (cond && !it.hasNext()))) {
                System.out.printf("Suspicious: %s::%s line %d-%d\n",
                        analysis.getKlazzName().replace("/", "."),
                        analysis.getMethodName(),
                        currentRangeBegin,
                        currentRangeLast);
                currentRangeBegin = currentRangeLast = -1;
            }
        }
    }
}

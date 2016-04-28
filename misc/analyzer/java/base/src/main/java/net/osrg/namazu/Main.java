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

import org.kohsuke.args4j.Argument;
import org.kohsuke.args4j.CmdLineException;
import org.kohsuke.args4j.CmdLineParser;
import org.kohsuke.args4j.Option;

import java.io.IOException;

public class Main {
    @Option(name = "-c", aliases = "--classes-path", metaVar = "classesPath", required = true, usage = "Java classes path")
    private String classesPath = null; // TODO: support multiple paths

    @Argument(index = 0, metaVar = "storagePath", required = true, usage = "Namazu storage directory path")
    private String storagePath = null;

    public static void main(String args[]) {
        Main inst = new Main();
        CmdLineParser parser = new CmdLineParser(inst);
        try {
            parser.parseArgument(args);
        } catch (CmdLineException e) {
            System.err.println(e.getMessage());
            parser.printUsage(System.err);
            System.exit(1);
        }

        Analyzer analyzer = new Analyzer(inst.storagePath, inst.classesPath);
        try {
            analyzer.analyze();
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
        AnalysisReporter reporter = new AnalysisReporter(analyzer);
        reporter.report();
    }
}

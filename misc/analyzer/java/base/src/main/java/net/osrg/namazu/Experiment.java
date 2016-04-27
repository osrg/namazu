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


import net.arnx.jsonic.JSON;
import org.apache.commons.io.FileUtils;
import org.apache.commons.io.filefilter.DirectoryFileFilter;
import org.apache.commons.io.filefilter.RegexFileFilter;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.jacoco.core.tools.ExecFileLoader;

import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;


public class Experiment {
    public final static String DEFAULT_NMZ_RESULT_JSON_PATH = "result.json";
    public final static String DEFAULT_JACOCO_PATH_REGEX = "jacoco.exec";
    static final Logger LOG = LogManager.getLogger(Experiment.class);
    /**
     * Namazu experiment dir (e.g. "0000002a")
     */
    private File dir;
    /**
     * Equivalent to (boolean) this.resultJsonMap["successful"]
     */
    private boolean successful;
    /**
     * Namazu result.json
     */
    private Map<String, Object> resultJsonMap;
    /**
     * JaCoCo execution file loader
     */
    private List<ExecFileLoader> execFileLoaders;
    /**
     * Pattern
     */
    private ExperimentPattern pattern;

    public Experiment(File dir) throws IOException {
        this(dir, DEFAULT_NMZ_RESULT_JSON_PATH, DEFAULT_JACOCO_PATH_REGEX);
    }

    public Experiment(File dir, String eqResultJsonPath, String jacocoPathRegex) throws IOException {
        this.dir = dir;
        this.resultJsonMap = new JSON().parse(new FileReader(new File(dir, eqResultJsonPath)));
        this.successful = (boolean) this.resultJsonMap.get("successful");

        this.execFileLoaders = new ArrayList<ExecFileLoader>();
        File[] jacocoFiles = FileUtils.listFiles(
                dir,
                new RegexFileFilter(jacocoPathRegex),
                DirectoryFileFilter.DIRECTORY).toArray(new File[0]);
        Arrays.sort(jacocoFiles);
        for (File jacocoFile : jacocoFiles) {
            ExecFileLoader loader = new ExecFileLoader();
            // LOG.debug("Loading {}", jacocoFile);
            loader.load(jacocoFile);
            this.execFileLoaders.add(loader);
        }

        this.pattern = new ExperimentPattern();
    }

    public File getDir() {
        return dir;
    }

    public boolean isSuccessful() {
        return successful;
    }

    public Map<String, Object> getResultJsonMap() {
        return resultJsonMap;
    }

    public List<ExecFileLoader> getExecFileLoaders() {
        return execFileLoaders;
    }

    public ExperimentPattern getPattern() {
        return pattern;
    }
}

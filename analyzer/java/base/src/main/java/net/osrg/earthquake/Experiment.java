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


import net.arnx.jsonic.JSON;
import org.jacoco.core.tools.ExecFileLoader;

import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.Map;

public class Experiment {
    public final static String DEFAULT_EQ_RESULT_JSON_PATH = "result.json";
    public final static String DEFAULT_JACOCO_EXEC_PATH = "jacoco/jacoco.exec"; // TODO: support Windows path sep?

    public Experiment(String dirPath) throws IOException {
        this(dirPath, DEFAULT_EQ_RESULT_JSON_PATH, DEFAULT_JACOCO_EXEC_PATH);
    }

    public Experiment(String dirPath, String eqResultJsonPath, String jacocoExecPath) throws IOException {
        this.dirPath = dirPath;
        JSON json = new JSON();
        this.resultJsonMap = json.parse(new FileReader(new File(dirPath, eqResultJsonPath)));
        this.successful = (boolean) this.resultJsonMap.get("successful");
        this.execFileLoader = new ExecFileLoader();
        this.execFileLoader.load(new File(dirPath, jacocoExecPath));
    }

    /**
     * Earthquake experiment dir path (e.g. "0000002a")
     */
    private String dirPath;

    /**
     * Equivalent to (boolean) this.resultJsonMap["successful"]
     */
    private boolean successful;

    /**
     * Earthquake result.json
     */
    private Map<String, Object> resultJsonMap;

    /**
     * JaCoCo execution file loader
     */
    private ExecFileLoader execFileLoader;

    public String getDirPath() {
        return dirPath;
    }

    public boolean isSuccessful() {
        return successful;
    }

    public Map<String, Object> getResultJsonMap() {
        return resultJsonMap;
    }

    public ExecFileLoader getExecFileLoader() {
        return execFileLoader;
    }

}

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

import com.google.common.hash.HashCode;
import com.google.common.hash.Hasher;
import com.google.common.hash.Hashing;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.nio.charset.Charset;

public class ExperimentPattern {
    static final Logger LOG = LogManager.getLogger(ExperimentPattern.class);
    Hasher hasher;

    public ExperimentPattern() {
        this.hasher = Hashing.sha256().newHasher();
    }

    // TODO: pluggable
    private int normalizeCount(String rowKey, String colKey, int count) {
        return count > 0 ? 1 : 0;
        // return count;
    }

    public void putCount(String klazzName, String methodName, int line, String nodeID, boolean succ, int count) {
        String rowKey = String.format("%s:%s:%d", klazzName, methodName, line);
        String colKey = nodeID;
        int normalizedCount = this.normalizeCount(rowKey, colKey, count);
        String s = String.format("%s\t%s\t%d\n", rowKey, colKey, normalizedCount);
        this.hasher.putString(s, Charset.defaultCharset());
    }

    public HashCode computeHash() {
        return this.hasher.hash();
    }
}

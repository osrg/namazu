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

package org.osrg.earthquake;

import org.osrg.earthquake.*;
import java.util.*;
import org.jboss.byteman.rule.*;
import org.jboss.byteman.rule.helper.*;

public class JSONEQHelper extends Helper
{
    static JSONInspector inspector;

    static {
	inspector = new JSONInspector();
    };

    public JSONEQHelper(Rule rule) {
	super(rule);
    }

    public static void activated() {
	inspector.Initiation();
	// System.out.println("BTM: initiation to orchestrator completed");
    }

    public static void deactivated() {
        //stopThread();
    }

    public void eventFuncCall(String name) {
	inspector.EventFuncCall(name);
    }

    public void eventFuncCall(String name, Map<String, Object> argMap) {
	inspector.EventFuncCall(name, argMap);
    }

    public void stopInspection() {
	inspector.StopInspection();
    }
}



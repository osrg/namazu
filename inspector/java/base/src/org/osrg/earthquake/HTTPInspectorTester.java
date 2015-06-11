// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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
import java.util.*;

public class HTTPInspectorTester {
    //TODO: introduce JUnit unittest
    public static void main(String args[]) {
        Inspector i = new HTTPInspector();
        i.Initiation();

        System.out.println("sending funcCall event (testMethod)");
        i.EventFuncCall("testMethod");
        System.out.println("sent funcCall event (testMethod)");

	Map<String, Object> argMap = new HashMap<String, Object>();
	argMap.put("foo", "bar");
	argMap.put("n", 42);
        System.out.println("sending funcCall event (testMethod2)");
        i.EventFuncCall("testMethod2", argMap);
        System.out.println("sent funcCall event (testMethod2)");
	
	i.StopInspection();
    }
}

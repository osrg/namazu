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

import java.util.*;
import net.arnx.jsonic.*;

public class FunctionCallEvent extends Event{
    public FunctionCallEvent(String process, String functionName, Map<String, Object> argMap){
	this.type = "event";
	this.klazz = "FunctionCallEvent";
	this.process = process;
	this.uuid = this.generateUUID();
	this.deferred = true;
	this.option = new HashMap<String, Object>();
	this.option.put("func_name", (Object)functionName);
	if (argMap != null) {
	    for (Map.Entry<String, Object> entry : argMap.entrySet()) {
		String key = entry.getKey();
		if ( key.equals("func_name") ) {
		    throw new RuntimeException("Do not specify func_name here");
		}
		Object val = entry.getValue();
		this.option.put(key, val);
	    }
	}
    }
}


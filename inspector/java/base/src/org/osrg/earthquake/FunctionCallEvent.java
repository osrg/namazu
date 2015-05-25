package org.osrg.earthquake;

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


// Copyright (C) 2014 - 2015 Nippon Telegraph and Telephone Corporation.
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


import java.io.*;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.SynchronousQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.ReentrantLock;
import java.util.logging.*;
import java.net.*;

import sun.misc.Signal;
import sun.misc.SignalHandler;

import org.apache.http.*;
import org.apache.http.client.*;
import org.apache.http.client.fluent.*;
import org.apache.http.entity.*;

import net.arnx.jsonic.*;

public class JSONInspector implements Inspector {
    private boolean Disabled = false;
    private Logger LOGGER;

    
    private String ProcessID;
    private String OrchestratorURL = "http://localhost:10000";


    private Map<String, SynchronousQueue<Object>> waitingMap;

    
    public JSONInspector() {
        LOGGER = Logger.getLogger(this.getClass().getName());
        LOGGER.setLevel(Level.INFO);

        try {
            FileHandler logFileHandler = new FileHandler("/tmp/earthquake-inspection-java.log");
            logFileHandler.setFormatter(new SimpleFormatter());
            LOGGER.addHandler(logFileHandler);
        } catch (IOException e) {
            System.err.println("failed to initialize file hander for logging: " + e);
            System.exit(1);
        }

        String _Disabled = System.getenv("EQ_DISABLE");
        if (_Disabled != null) {
            LOGGER.info("inspection is disabled");
            Disabled = true;
            return;
        }

        if (System.getenv("EQ_MODE_DIRECT") != null) {
            LOGGER.warning("direct/non-direct mode has been abandoned. you can use HTTP proxy (over UNIX domain socket, with socat/netcat/..) instead of non-direct (proxied) mode.");
        }
        if (System.getenv("EQ_GA_TCP_PORT") != null) {
	    LOGGER.warning("EQ_GA_TCP_PORT has been abandoned. you can set EQ_ORCHESTRATOR_URL instad.");
        }


        if (System.getenv("EQ_ORCHESTRATOR_URL") == null) {
            LOGGER.warning("orchestrator url not given (EQ_ORCHESTRATOR_URL), default=" + OrchestratorURL);
        } else {
	    OrchestratorURL = System.getenv("EQ_ORCHESTRATOR_URL");
	}
	
        ProcessID = System.getenv("EQ_ENV_PROCESS_ID");
        if (ProcessID == null) {
            LOGGER.severe("process id required but not given (EQ_ENV_PROCESS_ID");
            System.exit(1);
        }
        LOGGER.info("Process ID: " + ProcessID);

	waitingMap = new HashMap<String, SynchronousQueue<Object>>();

        Signal signal = new Signal("TERM");
        Signal.handle(signal, new SignalHandler() {
		public void handle(Signal signal) {
		    LOGGER.info("singal: " + signal + " catched");
		    if ( reader != null ) { reader.kill(); }
		    System.exit(0);
		}
	    });
    }

    private boolean running = true;

    private class ReaderThread extends Thread {
        public void kill() {
	    // TODO?: send InspectionEndEvent?
	    running = false;
        }
        public void run() {
            LOGGER.info("reader thread starts");
	    String url = OrchestratorURL + "/api/v1/" + ProcessID;
	    int errorCount = 0;
            while (running) {
                LOGGER.fine("reader thread loop");
		try{
		    LOGGER.info("GET ==> " + url);		    
		    String contentStr =
			Request.Get(url).execute().returnContent().asString();
		    LOGGER.info("GET <== " + url + ": result=" + contentStr);
		    onGotActionJson(contentStr);		    
		    errorCount = 0;
		} catch (Exception ex){
		    LOGGER.severe("reader thread: "+ ex.toString());
		    errorCount++;
		}
		if (errorCount > 0 ){
		    LOGGER.warning("sleeping, errorCount=" + errorCount);
		    try{
			Thread.sleep(errorCount * 1000);
		    }catch(InterruptedException ie){}
		}
            } // while running loop
        }
    }

    private ReaderThread reader;

    public void Initiation() {
        if (Disabled) {
            return;
        }
        reader = new ReaderThread();
        reader.start();
    }


    private void postEventJson(String jsonStr){
	String url = OrchestratorURL + "/api/v1";
	try {
	    LOGGER.info("POST ==>" + url + ":request=" + jsonStr);	    
	    String contentStr = Request.Post(url)
		.bodyString(jsonStr, ContentType.APPLICATION_JSON)
		.execute().returnContent().asString();
	    LOGGER.info("POST <==" + url + ":result=" + contentStr);
	} catch (Exception ex) {
	    LOGGER.severe("postEventJson: " + ex.toString());
	    ex.printStackTrace();
	}
    }

    private void onGotActionJson(String jsonStr){
	LOGGER.info("got Action (unparsed)" + jsonStr);	
	Entity act = JSON.decode(jsonStr, Entity.class);
	LOGGER.info("got Action (parsed)" + act.toString());
	if ( act.klazz.equals("PassDeferredEventAction") ) {
	    String eventUuid = (String)act.option.get("event_uuid");
	    notifyPassDeferredEventAction(eventUuid);
	} else {
	    LOGGER.severe("Unknown action class " + act.klazz);
	}
    }

    public void EventFuncCall(String funcName) {
	EventFuncCall(funcName, null);
    }
	
    public void EventFuncCall(String funcName, Map<String, Object> argMap) {
	//TODO: create a thread on each of EventFuncCall
        if (Disabled) {
            LOGGER.fine("already disabled");
            return;
        }

        if (!running) {
            LOGGER.fine("killed");
            return;
        }
	LOGGER.finest("EventFuncCall(): funcName=" + funcName);
	FunctionCallEvent ev = new FunctionCallEvent(ProcessID, funcName, argMap);
	String jsonStr = JSON.encode(ev, true);
        LOGGER.finest("FunctionCallEvent JSON " + jsonStr);
	postEventJson(jsonStr);	
	waitForPassDeferredEventAction(ev.uuid);
    }

    private void waitForPassDeferredEventAction(String eventUuid){
	SynchronousQueue<Object> q = new SynchronousQueue<Object>();
	synchronized (waitingMap) {
	    waitingMap.put(eventUuid, q);
	}
	try {
	    LOGGER.info("waitForPassDeferredEventAction ==> wait enter eventUuid=" + eventUuid);
	    q.take();
	    LOGGER.info("waitForPassDeferredEventAction <== wait leave eventUuid=" + eventUuid);	    
	} catch (InterruptedException e) {
	    LOGGER.severe("interrupted: " + e);
	    System.exit(1); // TODO: handling
	}
    }

    private void notifyPassDeferredEventAction(String eventUuid){
	synchronized (waitingMap) {
	    SynchronousQueue<Object> q = waitingMap.get(eventUuid);
	    Object token = new Object();
	    q.offer(token);
	    waitingMap.remove(q);
	}
    }

    public void StopInspection() {
        if ( reader != null ) { reader.kill(); }
    }
}

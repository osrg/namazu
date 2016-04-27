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

package net.osrg.namazu;

import com.google.protobuf.InvalidProtocolBufferException;
import sun.misc.Signal;
import sun.misc.SignalHandler;

import java.io.*;
import java.net.Socket;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.SynchronousQueue;
import java.util.logging.FileHandler;
import java.util.logging.Level;
import java.util.logging.Logger;
import java.util.logging.SimpleFormatter;

public class PBInspector implements Inspector {
    private boolean Direct = false;
    private boolean Disabled = false;
    private boolean NoInitiation = false;
    private boolean Dryrun = false;
    private String EntityID;
    private int GATCPPort = 10000;

    private Logger LOGGER;

    private Socket GASock;
    private DataOutputStream GAOutstream;
    private DataInputStream GAInstream;

    private Map<Integer, SynchronousQueue<Object>> waitingMap;

    private int SendReq(InspectorMessage.InspectorMsgReq req) {
        return SendReq(GAOutstream, req);
    }

    private int SendReq(DataOutputStream outstream, InspectorMessage.InspectorMsgReq req) {
        byte[] serialized = req.toByteArray();
        byte[] lengthBuf = ByteBuffer.allocate(4).order(ByteOrder.LITTLE_ENDIAN).putInt(serialized.length).array();

        try {
            outstream.write(lengthBuf);
            outstream.write(serialized);
        } catch (IOException e) {
            return 1;
        }

        return 0;
    }

    private InspectorMessage.InspectorMsgRsp RecvRsp() {
        return RecvRsp(GAInstream);
    }

    private InspectorMessage.InspectorMsgRsp RecvRsp(DataInputStream instream) {
        byte[] lengthBuf = new byte[4];

        try {
            instream.read(lengthBuf, 0, 4);
        } catch (IOException e) {
            LOGGER.severe("failed to read header of response: " + e);
            return null;
        }

        int length = ByteBuffer.wrap(lengthBuf).order(ByteOrder.LITTLE_ENDIAN).getInt();
        byte[] rspBuf = new byte[length];

        try {
            instream.read(rspBuf, 0, length);
        } catch (IOException e) {
            LOGGER.severe("failed to read body of response: " + e);
            return null;
        }

        InspectorMessage.InspectorMsgRsp rsp;
        try {
            rsp = InspectorMessage.InspectorMsgRsp.parseFrom(rspBuf);
        } catch (InvalidProtocolBufferException e) {
            LOGGER.severe("failed to parse response: " + e);
            return null;
        }

        return rsp;
    }

    private InspectorMessage.InspectorMsgRsp ExecReq(InspectorMessage.InspectorMsgReq req) {
        int ret = SendReq(req);
        if (ret != 0) {
            LOGGER.severe("failed to send request");
            System.exit(1);
        }

        InspectorMessage.InspectorMsgRsp rsp = RecvRsp();
        if (rsp == null) {
            LOGGER.severe("failed to receive response");
            System.exit(1);
        }

        return rsp;
    }

    public PBInspector() {
        LOGGER = Logger.getLogger(this.getClass().getName());
        LOGGER.setLevel(Level.INFO);

        Signal signal = new Signal("TERM");
        Signal.handle(signal, new SignalHandler() {
            public void handle(Signal signal) {
                LOGGER.info("singal: " + signal + " catched");
                if (reader != null) {
                    reader.kill();
                }

                System.exit(0);
            }
        });

        try {
            FileHandler logFileHandler = new FileHandler("/tmp/namazu-inspection-java.log");
            logFileHandler.setFormatter(new SimpleFormatter());
            LOGGER.addHandler(logFileHandler);
        } catch (IOException e) {
            System.err.println("failed to initialize file hander for logging: " + e);
            System.exit(1);
        }

        String _Disabled = System.getenv("NMZ_DISABLE");
        if (_Disabled != null) {
            LOGGER.info("inspection is disabled");
            Disabled = true;
            return;
        }

        String _Dryrun = System.getenv("NMZ_DRYRUN");
        if (_Dryrun != null) {
            LOGGER.info("inspection in dryrun");
            Dryrun = true;
        }

        String _NoInitiation = System.getenv("NMZ_NO_INITIATION");
        if (_NoInitiation != null) {
            LOGGER.info("no initiation, connection per thread model");
            NoInitiation = true;
        }

        String _Direct = System.getenv("NMZ_MODE_DIRECT");
        if (_Direct != null) {
            LOGGER.info("run in direct mode");
            Direct = true;
        } else {
            LOGGER.info("run in non direct mode");
        }

        EntityID = System.getenv("NMZ_ENV_ENTITY_ID");
        if (EntityID == null) {
            LOGGER.severe("entity id required but not given (NMZ_ENV_ENTITY_ID");
            System.exit(1);
        }
        LOGGER.info("Entity ID: " + EntityID);

        String _GATCPPort = System.getenv("NMZ_GA_TCP_PORT");
        if (_GATCPPort != null) {
            GATCPPort = Integer.parseInt(_GATCPPort);
            LOGGER.info("given TCP port of guest agent: " + GATCPPort);
        }

        waitingMap = new HashMap<Integer, SynchronousQueue<Object>>();
    }

    private boolean running = true;

    private class ReaderThread extends Thread {

        public void kill() {
            InspectorMessage.InspectorMsgReq_Event_Exit.Builder evExitBuilder = InspectorMessage.InspectorMsgReq_Event_Exit.newBuilder();
            InspectorMessage.InspectorMsgReq_Event_Exit evExit = evExitBuilder.setExitCode(0).build(); // TODO: exit code

            InspectorMessage.InspectorMsgReq_Event.Builder evBuilder = InspectorMessage.InspectorMsgReq_Event.newBuilder();
            InspectorMessage.InspectorMsgReq_Event ev = evBuilder
                    .setType(InspectorMessage.InspectorMsgReq_Event.Type.EXIT)
                    .setExit(evExit).build();

            running = false;
            sendEvent(ev, false, null, null);

            try {
                GAInstream.close();
            } catch (IOException e) {
                LOGGER.severe("closing GAInstream failed: " + e);
            }
        }

        public void run() {
            LOGGER.info("reader thread starts");

            while (running) {
                LOGGER.fine("reader thread loop");

                InspectorMessage.InspectorMsgRsp rsp = RecvRsp();
                if (rsp == null) {
                    // TODO: need to determine orchestrator is broken or kill() is called
                    if (!running) {
                        LOGGER.info("exiting reader thread");
                        return;
                    }
                }

                if (rsp.getRes() == InspectorMessage.InspectorMsgRsp.Result.END) {
                    LOGGER.info("inspection end");
                    Disabled = true;
                    break;
                }

                if (rsp.getRes() != InspectorMessage.InspectorMsgRsp.Result.ACK) {
                    LOGGER.severe("invalid response: " + rsp.getRes());
                    System.exit(1);
                }

                int msgID = rsp.getMsgId();
                LOGGER.info("recieved response, message ID: " + Integer.toString(msgID));
                synchronized (waitingMap) {
                    SynchronousQueue<Object> q = waitingMap.get(msgID);
                    Object token = new Object();
                    q.offer(token);
                    waitingMap.remove(q);
                }
            }
        }
    }

    private ReaderThread reader;

    public void Initiation() {
        if (Disabled) {
            return;
        }

        if (NoInitiation) {
            LOGGER.info("no initiation mode");
            return;
        }

        if (!Dryrun) {
            try {
                GASock = new Socket("localhost", GATCPPort);

                OutputStream out = GASock.getOutputStream();
                GAOutstream = new DataOutputStream(out);

                InputStream in = GASock.getInputStream();
                GAInstream = new DataInputStream(in);
            } catch (IOException e) {
                LOGGER.severe("failed to connect to guest agent: " + e);
                System.exit(1);
            }
        }

        InspectorMessage.InspectorMsgReq_Initiation.Builder initiationReqBuilder = InspectorMessage.InspectorMsgReq_Initiation.newBuilder();
        InspectorMessage.InspectorMsgReq_Initiation initiationReq = initiationReqBuilder.setEntityId(EntityID).build();

        InspectorMessage.InspectorMsgReq.Builder reqBuilder = InspectorMessage.InspectorMsgReq.newBuilder();
        InspectorMessage.InspectorMsgReq req = reqBuilder.setPid(0 /* FIXME */)
                .setTid((int) Thread.currentThread().getId())
                .setType(InspectorMessage.InspectorMsgReq.Type.INITIATION)
                .setMsgId(0)
                .setEntityId(EntityID)
                .setInitiation(initiationReq).build();

        if (Dryrun) {
            // TODO: dump initiation message
            System.out.println("initiation message: " + req.toString());
            return;
        }

        LOGGER.info("executing request for initiation");
        InspectorMessage.InspectorMsgRsp rsp = ExecReq(req);
        if (rsp.getRes() != InspectorMessage.InspectorMsgRsp.Result.ACK) {
            LOGGER.severe("initiation failed, result: " + rsp.getRes());
            System.exit(1);
        }

        LOGGER.info("initiation succeed");

        reader = new ReaderThread();
        reader.start();
    }

    private int MsgID = 1;

    private synchronized int nextMsgID() {
        int ret;
        ret = MsgID;
        MsgID++;
        return ret;
    }

    private void sendEvent(InspectorMessage.InspectorMsgReq_Event ev, boolean needRsp,
                           InspectorMessage.InspectorMsgReq_JavaSpecificFields_StackTraceElement traces[],
                           InspectorMessage.InspectorMsgReq_JavaSpecificFields_Params[] params)  {
        int msgID = nextMsgID();

        InspectorMessage.InspectorMsgReq_JavaSpecificFields.Builder javaSpecificFieldBuilder =
                InspectorMessage.InspectorMsgReq_JavaSpecificFields.newBuilder();
        javaSpecificFieldBuilder.setThreadName(Thread.currentThread().getName());

        if (traces == null) {
            javaSpecificFieldBuilder.setNrStackTraceElements(0);
        } else {
            javaSpecificFieldBuilder.setNrStackTraceElements(traces.length);
            for (int i = 0; i < traces.length; i++) {
                javaSpecificFieldBuilder.addStackTraceElements(traces[i]);
            }
        }

        if (params == null) {
            javaSpecificFieldBuilder.setNrParams(0);
        } else {
            javaSpecificFieldBuilder.setNrParams(params.length);
            for (int i = 0; i < params.length; i++) {
                javaSpecificFieldBuilder.addParams(params[i]);
            }
        }

        InspectorMessage.InspectorMsgReq_JavaSpecificFields javaSpecificField = javaSpecificFieldBuilder.build();

        InspectorMessage.InspectorMsgReq.Builder reqBuilder = InspectorMessage.InspectorMsgReq.newBuilder();
        InspectorMessage.InspectorMsgReq req = reqBuilder.setPid(0 /*FIXME*/)
                .setTid((int) Thread.currentThread().getId())
                .setType(InspectorMessage.InspectorMsgReq.Type.EVENT)
                .setMsgId(msgID)
                .setEntityId(EntityID)
                .setHasJavaSpecificFields(1)
                .setJavaSpecificFields(javaSpecificField)
                .setEvent(ev).build();

        if (Dryrun) {
            // TODO: dump message
            System.out.println("dryrun mode, do nothing");
            System.out.println("event message: " + req.toString());
            return;
        }

        if (NoInitiation) {
            final ThreadLocal<Socket> tlsSocket = new ThreadLocal<Socket>() {
                protected Socket initialValue() {
                    Socket sock = null;

                    try {
                        sock = new Socket("localhost", GATCPPort);
                        sock.setSoTimeout(0);
                    } catch (IOException e) {
                        LOGGER.severe("failed to connect to guest agent: " + e);
                        return null;
                    }

                    return sock;
                }
            };

            ThreadLocal<DataOutputStream> tlsOutputStream = new ThreadLocal<DataOutputStream>() {
                protected DataOutputStream initialValue() {
                    DataOutputStream stream = null;
                    if (tlsSocket == null) {
                        LOGGER.severe("tlsSocket is null");
                        return null;
                    }
                    Socket sock = tlsSocket.get();
                    if (sock == null) {
                        LOGGER.severe("sock is null");
                        return null;
                    }

                    try {
                        OutputStream out = sock.getOutputStream();
                        stream = new DataOutputStream(out);
                    } catch (IOException e) {
                        LOGGER.severe("failed to get DataOutputStream: " + e);
                        return null;
                    }

                    return stream;
                }
            };

            ThreadLocal<DataInputStream> tlsInputStream = new ThreadLocal<DataInputStream>() {
                protected DataInputStream initialValue() {
                    DataInputStream stream = null;
                     if (tlsSocket == null) {
                        LOGGER.severe("tlsSocket is null");
                        return null;
                    }
                    Socket sock = tlsSocket.get();
                    if (sock == null) {
                        LOGGER.severe("sock is null");
                        return null;
                    }

                     try {
                        InputStream in = sock.getInputStream();
                        stream = new DataInputStream(in);
                    } catch (IOException e) {
                        LOGGER.severe("failed to get DataOutputStream: " + e);
                        return null;
                    }

                    return stream;
                }
            };

            if (tlsInputStream != null && tlsOutputStream.get() != null) {
                SendReq(tlsOutputStream.get(), req);
                InspectorMessage.InspectorMsgRsp rsp = RecvRsp(tlsInputStream.get());
                if (rsp != null) {
                    LOGGER.fine("response message: " + rsp.toString());
                }
            } else {
                LOGGER.warning("socket ins't ready");
            }
        } else {
            SendReq(req);

            SynchronousQueue<Object> q = new SynchronousQueue<Object>();
            synchronized (waitingMap) {
                waitingMap.put(msgID, q);
            }

            if (!needRsp) {
                return;
            }

            try {
                q.take();
            } catch (InterruptedException e) {
                LOGGER.severe("interrupted: " + e);
                System.exit(1); // TODO: handling
            }
        }
    }

    private InspectorMessage.InspectorMsgReq_JavaSpecificFields_StackTraceElement[] makeStackTrace() {
        StackTraceElement traces[] = Thread.currentThread().getStackTrace();

        /* CAUTION: heuristics */
        int maxRuleJavaIdx = -1;
        for (int i = 0; i < traces.length; i++) {
            StackTraceElement trace = traces[i];

            if (trace == null) {
                continue;
            }

            String fileName = trace.getFileName();

            if (fileName == null) {
                continue;
            }

            if (fileName.equals("Rule.java")) {
                maxRuleJavaIdx = i;
            }
        }

        if (maxRuleJavaIdx == -1) {
            LOGGER.severe("unexpected call stack");
            System.exit(1);
        }

        InspectorMessage.InspectorMsgReq_JavaSpecificFields_StackTraceElement[] ret = new InspectorMessage.InspectorMsgReq_JavaSpecificFields_StackTraceElement[traces.length - (maxRuleJavaIdx + 1)];
        for (int i = maxRuleJavaIdx + 1, j = 0; i < traces.length; i++, j++) {
            StackTraceElement trace = traces[i];

            if (trace == null) {
                LOGGER.info("stack trace entry %d is null" + i);
                continue;
            }

            InspectorMessage.InspectorMsgReq_JavaSpecificFields_StackTraceElement.Builder traceBuilder = InspectorMessage.InspectorMsgReq_JavaSpecificFields_StackTraceElement.newBuilder();
            traceBuilder.setClassName(trace.getClassName() != null ? trace.getClassName() : "<no class name>")
                    .setFileName(trace.getFileName() != null ? trace.getFileName() : "<no file name>")
                    .setMethodName(trace.getMethodName() != null ? trace.getMethodName() : "<no method name>")
                    .setLineNumber(trace.getLineNumber());

            InspectorMessage.InspectorMsgReq_JavaSpecificFields_StackTraceElement newElement = traceBuilder.build();

            ret[j] = newElement;
        }

        return ret;
    }

    public void EventFuncCall(String funcName) {
        if (Disabled) {
            LOGGER.fine("already disabled");
            return;
        }

        if (!running) {
            LOGGER.fine("killed");
            return;
        }

        LOGGER.finest("EventFuncCall: " + funcName);
        InspectorMessage.InspectorMsgReq_Event_FuncCall.Builder evFunBuilder = InspectorMessage.InspectorMsgReq_Event_FuncCall.newBuilder();
        InspectorMessage.InspectorMsgReq_Event_FuncCall evFun = evFunBuilder.setName(funcName).build();

        InspectorMessage.InspectorMsgReq_Event.Builder evBuilder = InspectorMessage.InspectorMsgReq_Event.newBuilder();
        InspectorMessage.InspectorMsgReq_Event ev = evBuilder
                .setType(InspectorMessage.InspectorMsgReq_Event.Type.FUNC_CALL)
                .setFuncCall(evFun).build();

        sendEvent(ev, true, makeStackTrace(), null);
    }

    public void EventFuncReturn(String funcName) {
        if (Disabled) {
            LOGGER.fine("already disabled");
            return;
        }

        if (!running) {
            LOGGER.fine("killed");
            return;
        }

        LOGGER.finest("EventFuncReturn: " + funcName);
        InspectorMessage.InspectorMsgReq_Event_FuncReturn.Builder evFunBuilder = InspectorMessage.InspectorMsgReq_Event_FuncReturn.newBuilder();
        InspectorMessage.InspectorMsgReq_Event_FuncReturn evFun = evFunBuilder.setName(funcName).build();

        InspectorMessage.InspectorMsgReq_Event.Builder evBuilder = InspectorMessage.InspectorMsgReq_Event.newBuilder();
        InspectorMessage.InspectorMsgReq_Event ev = evBuilder
                .setType(InspectorMessage.InspectorMsgReq_Event.Type.FUNC_RETURN)
                .setFuncReturn(evFun).build();

        sendEvent(ev, true, makeStackTrace(), null);
    }

    private InspectorMessage.InspectorMsgReq_JavaSpecificFields_Params[] makeParamsArray(Map<String, Object> paramMap) {
        InspectorMessage.InspectorMsgReq_JavaSpecificFields_Params[] ret;
        ret = new InspectorMessage.InspectorMsgReq_JavaSpecificFields_Params[paramMap.size()];

        int i = 0;
        for (Map.Entry<String, Object> e: paramMap.entrySet()) {
            InspectorMessage.InspectorMsgReq_JavaSpecificFields_Params.Builder paramBuilder = InspectorMessage.InspectorMsgReq_JavaSpecificFields_Params.newBuilder();

            ret[i++] = paramBuilder.setName(e.getKey()).setValue(e.getValue() == null ? "null" : e.getValue().toString()).build();
        }

        return ret;
    }

    private boolean classFilterMatch(String className) {
        StackTraceElement traces[] = Thread.currentThread().getStackTrace();

        for (int i = 0; i < traces.length; i++) {
            StackTraceElement trace = traces[i];

            String name = trace.getClassName();
            if (name == null) {
                continue;
            }

            if (name.equals(className)) {
                LOGGER.info("class filter match: " + className);
                return true;
            }
        }

        LOGGER.info("class filter dones't match: " + className);
        return false;
    }

    public void EventFuncCall(String funcName, String classFilter) {
        if (!classFilterMatch(classFilter)) {
            return;
        }

        EventFuncCall(funcName);
    }

    public void EventFuncReturn(String funcName, String classFilter) {
        if (!classFilterMatch(classFilter)) {
            return;
        }

        EventFuncReturn(funcName);
    }

    public void EventFuncCall(String funcName, Map<String, Object> paramMap) {
         if (Disabled) {
            LOGGER.fine("already disabled");
            return;
        }

        if (!running) {
            LOGGER.fine("killed");
            return;
        }

        LOGGER.finest("EventFuncCall: " + funcName);
        LOGGER.info("paramMap: " + paramMap.toString());
        InspectorMessage.InspectorMsgReq_Event_FuncCall.Builder evFunBuilder = InspectorMessage.InspectorMsgReq_Event_FuncCall.newBuilder();
        InspectorMessage.InspectorMsgReq_Event_FuncCall evFun = evFunBuilder.setName(funcName).build();

        InspectorMessage.InspectorMsgReq_Event.Builder evBuilder = InspectorMessage.InspectorMsgReq_Event.newBuilder();
        InspectorMessage.InspectorMsgReq_Event ev = evBuilder
                .setType(InspectorMessage.InspectorMsgReq_Event.Type.FUNC_CALL)
                .setFuncCall(evFun).build();

        sendEvent(ev, true, makeStackTrace(), makeParamsArray(paramMap));
    }

    public void EventFuncReturn(String funcName, Map<String, Object> paramMap) {
         if (Disabled) {
            LOGGER.fine("already disabled");
            return;
        }

        if (!running) {
            LOGGER.fine("killed");
            return;
        }

        LOGGER.finest("EventFuncReturn: " + funcName);
        LOGGER.info("paramMap: " + paramMap.toString());
        InspectorMessage.InspectorMsgReq_Event_FuncReturn.Builder evFunBuilder = InspectorMessage.InspectorMsgReq_Event_FuncReturn.newBuilder();
        InspectorMessage.InspectorMsgReq_Event_FuncReturn evFun = evFunBuilder.setName(funcName).build();

        InspectorMessage.InspectorMsgReq_Event.Builder evBuilder = InspectorMessage.InspectorMsgReq_Event.newBuilder();
        InspectorMessage.InspectorMsgReq_Event ev = evBuilder
                .setType(InspectorMessage.InspectorMsgReq_Event.Type.FUNC_RETURN)
                .setFuncReturn(evFun).build();

        sendEvent(ev, true, makeStackTrace(), makeParamsArray(paramMap));
    }

    public void EventFuncCall(String funcName, Map<String, Object> paramMap, String classFilter) {
        if (!classFilterMatch(classFilter)) {
            return;
        }

        EventFuncCall(funcName, paramMap);
    }

    public void EventFuncReturn(String funcName, Map<String, Object> paramMap, String classFilter) {
        if (!classFilterMatch(classFilter)) {
            return;
        }

        EventFuncReturn(funcName, paramMap);
    }

    public void StopInspection() {
        if (Dryrun) {
            System.out.println("dryrun mode, do nothing");
            return;
        }

        if (reader != null) {
            reader.kill();
        }
    }
}

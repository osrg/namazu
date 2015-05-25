package org.osrg.earthquake;

import java.util.*;
import net.arnx.jsonic.*;

public class Entity{
    public String type;
    @JSONHint(name="class")
    public String klazz;
    public String process;
    public String uuid;
    public Map<String, Object> option;
    protected static String generateUUID() {
	return UUID.randomUUID().toString();
    }

    @Override
    public String toString(){
	return String.format("<Entity type=%s, klazz=%s, process=%s, uuid=%s, ..>", type, klazz, process, uuid);
    }
}

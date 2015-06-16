/**
 * Created by mitake on 4/20/15.
 * Original: https://github.com/osrg/zookeeper/commits/mitake/readPacket
 */

import org.apache.zookeeper.*;
import org.apache.zookeeper.ZooDefs.*;
import org.apache.zookeeper.client.FourLetterWordMain;
import org.apache.zookeeper.common.X509Exception.SSLContextException;
import org.apache.zookeeper.data.Stat;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class CheckZnodeZkCli {

    public static class MyWatcher implements Watcher {
        public void process(WatchedEvent event) {

        }
    }

    public static void main(String args[]) {
        if (args.length != 1) {
            System.err.println("usage: CheckZnodeZkCli <server address:serverport,server address:serverport ...>");
            System.exit(1);
        }
        
        HashMap<String, Map<String, String>> maps = new HashMap<String, Map<String, String>>();
        String[] servers = args[0].split(",");
        
        for (String server : servers) {
            List<String> childrens = new ArrayList<String>();
            
            MyWatcher watcher = new MyWatcher();
            ZooKeeper zk;
            try {
                zk = new ZooKeeper(server, 30000, new MyWatcher());

                childrens = zk.getChildren("/", watcher);
                
                for (String children : childrens) {
                    if (children == null) {
                        continue;
                    }
                    if (children.equals("zookeeper")) {
                        continue;
                    }
                   
                    String value = zk.getData("/" + children , watcher, new Stat()).toString();
                    HashMap<String, String> znodeMap = new HashMap<String, String>();
                    znodeMap.put(children, value);
                    maps.put(server, znodeMap);
                }
                    
                System.out.println(childrens);
            } catch (InterruptedException e) {
                System.err.println("InterruptedException: " + e);
                System.exit(1);
            } catch (KeeperException e) {
                System.err.println("KeeperException: " + e);
                System.exit(1);
            } catch (IOException e) {
                System.err.println("IOException: " + e);
                System.exit(1);
            }
        }
        
        Map<String, String> forDiff = new HashMap<String, String>();
        forDiff = maps.get(servers[0]);
        for (int i = 0 ; i < servers.length; i++) {
            if (i == 0) {
                continue;
            }
            if (!forDiff.keySet().equals(maps.get(servers[i]).keySet())) {
                System.out.println("znode key diffrent");
                System.exit(1);
            }
            if (!forDiff.values().equals(maps.get(servers[i]).values())) {
                System.out.println("znode value diffrent");
                System.exit(1);
            }
        }
    }
}
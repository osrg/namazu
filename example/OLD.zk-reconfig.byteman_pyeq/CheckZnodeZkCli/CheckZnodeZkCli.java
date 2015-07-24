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
import java.util.Map.Entry;
import java.util.Set;

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

        HashMap<String, Map<String, String>> serverZnodeMaps = new HashMap<String, Map<String, String>>();
        String[] servers = args[0].split(",");

        for (String server : servers) {
            List<String> childrens = new ArrayList<String>();

            MyWatcher watcher = new MyWatcher();

            try {
                ZooKeeper zk = new ZooKeeper(server, 30000, new MyWatcher());

                childrens = zk.getChildren("/", watcher);
                HashMap<String, String> znodesAndVlues = new HashMap<String, String>();

                for (String children : childrens) {
                    if (children == null ||
                        children.equals("zookeeper")) {
                        continue;
                    }

                    byte[] byteValue = zk.getData("/" + children , watcher, new Stat());
                    String value = new String(byteValue);

                    znodesAndVlues.put(children, value);
                }
                serverZnodeMaps.put(server, znodesAndVlues);

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

        String forDiffServerName = new String();
        Map<String, String> forDiffServerZnode = new HashMap<String, String>();

        for (Entry<String, Map<String, String>> forDiffServerZnodeEntry : serverZnodeMaps.entrySet()) {
            forDiffServerName = forDiffServerZnodeEntry.getKey();
            forDiffServerZnode = forDiffServerZnodeEntry.getValue();

            for (Entry<String, Map<String, String>> serverZnode : serverZnodeMaps.entrySet()) {
                if (forDiffServerName.equals(serverZnode.getKey())) {
                    continue;
                }

                for (String znodeName : serverZnode.getValue().keySet()) {
                    if (!forDiffServerZnode.containsKey(znodeName)) {
                        String error = String.format("znode(%s) of server(%s) is not exist in server(%s)", znodeName, serverZnode.getKey(), forDiffServerName);
                        System.out.println(error);
                        System.exit(1);
                    }
                }
                for (String znodeValue : serverZnode.getValue().values()) {
                    if (!forDiffServerZnode.containsValue(znodeValue)) {
                        String error = String.format("znodeValue(%s) of server(%s) is not exist in server(%s)", znodeValue, serverZnode.getKey(), forDiffServerName);
                        System.out.println(error);
                        System.exit(1);
                    }
                }
            }
        }
        System.out.println("Znodes check success!");
    }
}

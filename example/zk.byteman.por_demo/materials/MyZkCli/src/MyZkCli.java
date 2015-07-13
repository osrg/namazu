/**
 * Created by mitake on 4/20/15.
 * Original: https://github.com/osrg/zookeeper/commits/mitake/readPacket
 */

import org.apache.zookeeper.*;
import org.apache.zookeeper.ZooDefs.*;

import java.io.IOException;
import java.util.ArrayList;

public class MyZkCli {

    static private class zkClientThread extends Thread {

        private ZooKeeper zk;

        zkClientThread(String server) {
            try {
                zk = new ZooKeeper(server, 30000, new MyWatcher());
            } catch (IOException e) {
                System.err.println("IOException during creating ZooKeeper instance: " + e);
                System.exit(1);
            }
        }

        private boolean createZNode() {
             String testdata = "testdata";
            try {
                zk.create("/test-", testdata.getBytes(), Ids.OPEN_ACL_UNSAFE, CreateMode.PERSISTENT_SEQUENTIAL);
            } catch (KeeperException e) {
                System.err.println("KeeperException during creating sequential znode: " + e);
                System.err.println("zookeeper instance: " + zk.toString());
                return false;
            } catch (InterruptedException e) {
                System.err.println("InterruptedException during creating sequential znode: " + e);
                System.err.println("zookeeper instance: " + zk.toString());
                return false;
            }

            return true;
        }

        public void run() {
            int nrRetry = 100;
            for (int i = 0; i < nrRetry; i++) {
                if (createZNode()) {
                    System.out.println("succeed to create znode on " + zk.toString());
                    return;
                }
                System.out.println("failed to create znode on " + zk.toString());
            }
        }
    }

    public static class MyWatcher implements Watcher {
        public void process(WatchedEvent event) {

        }
    }

    public static void main(String args[]) {
        if (args.length != 3) {
            System.err.println("usage: myZkCli <zookeeper server 1> <zookeeper server 2> <zookeeper server 3>");
            System.exit(1);
        }

        ArrayList<zkClientThread> threads = new ArrayList<zkClientThread>();
        for (int i = 0; i < 3; i++) {
            zkClientThread newThread = new zkClientThread(args[i]);
            threads.add(newThread);

            newThread.run();
        }

        for (zkClientThread thread : threads) {
            try {
                thread.join();
            } catch (InterruptedException e) {
                System.err.println("InterruptedException during join: " + e);
                System.exit(1);
            }
        }
    }
}

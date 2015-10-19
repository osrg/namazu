/**
 * Created by mitake on 4/20/15.
 * Original: https://github.com/osrg/zookeeper/commits/mitake/readPacket
 */

import org.apache.zookeeper.*;
import org.apache.zookeeper.ZooDefs.*;

import java.io.IOException;

public class CreateZnodeZkCli {

    public static class MyWatcher implements Watcher {
        public void process(WatchedEvent event) {

        }
    }
    public static void main(String args[]) {
        if (args.length != 1) {
            System.err.println("usage: CreateZnodeZkCli <server address:server port>");
            System.exit(1);
        }

        try {
            ZooKeeper zk = new ZooKeeper(args[0], 30000, new MyWatcher());

            String testdata = "fubar";
            for (int i = 0; i < 2; i++) {
                zk.create("/test-", testdata.getBytes(), ZooDefs.Ids.OPEN_ACL_UNSAFE, CreateMode.PERSISTENT_SEQUENTIAL);
            }
        } catch (IOException e) {
            System.err.println("IOException: " + e);
            System.exit(1);
        } catch (KeeperException e) {
            System.err.println("KeeperException: " + e);
            System.exit(1);
        } catch (InterruptedException e) {
            System.err.println("InterruptedException: " + e);
            System.exit(1);
        }

    }
}


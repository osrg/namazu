/**
 * Created by mitake on 4/20/15.
 * Original: https://github.com/osrg/zookeeper/commits/mitake/readPacket
 */

import org.apache.zookeeper.*;
import org.apache.zookeeper.data.Stat;

import java.io.IOException;

public class AddNodeZkCli {

    public static class MyWatcher implements Watcher {
        public void process(WatchedEvent event) {

        }
    }

    public static void addServer(String connectServer, String addServer) throws IOException, KeeperException, InterruptedException {
        ZooKeeper zk = new ZooKeeper(connectServer, 30000, new MyWatcher());
            zk.reconfig(addServer, null, null, -1, new Stat());
    }

    public static void main(String args[]) {
        if (args.length != 2) {
            System.err.println("usage: ReconfigZkCli <connect server address:connect server port> <add node parameter>");
            System.exit(1);
        }

        try {
                addServer(args[0], args[1]);

        } catch (IOException e) {
            System.err.println("IOException: " + e);
            System.exit(1);
        } catch (KeeperException e) {
            System.err.println("KeeperException: " + e);
            System.exit(1);
        } catch (InterruptedException e) {
            System.err.println("InterruptedException: " + e);
            System.exit(1);
        } catch (NumberFormatException e) {
            System.err.println("NumberFormatException: " + e);
            System.exit(1);
        }
    }
}


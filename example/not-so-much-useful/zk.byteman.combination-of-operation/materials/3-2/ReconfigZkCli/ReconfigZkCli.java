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

public class ReconfigZkCli {

    public static class MyWatcher implements Watcher {
        public void process(WatchedEvent event) {

        }
    }

    private static String getMode(String server, String port) throws NumberFormatException, SSLContextException, IOException {
        String mode = new String();
        String stat = new String();
        stat = FourLetterWordMain.send4LetterWord(server, Integer.parseInt(port), "stat");

        if (stat.indexOf("leader") >= 0) {
            mode = "leader";
            System.out.println(server + ":" + port + " is leader");
        } else {
            mode = "follower";
            System.out.println(server + ":" + port + " is follower");
        }
        return mode;
    }

    private static void removeLeader(String server, String port) throws IOException, KeeperException, InterruptedException {
        MyWatcher watcher = new MyWatcher();
        ZooKeeper zk = new ZooKeeper(server, 30000, new MyWatcher());
        String leaderServer = new String();
        String configAll = new String(zk.getConfig(watcher, new Stat()));
        String[] configArray = configAll.split("\\n");
        String leaderServerId = null;
        for (String config : configArray) {
            if (config.indexOf(port) >= 0){
                leaderServer = config;
                String prefix = config.split("=")[0];
                int index = prefix.indexOf(".");
                leaderServerId = prefix.substring(index + 1);
                break;
            }
        }
//        File output = new File("/tmp/removedServer.id");
//        if (!output.exists()) {
//            FileOutputStream fos = new FileOutputStream(output);
//            OutputStreamWriter osw = new OutputStreamWriter(fos);
//            PrintWriter pw = new PrintWriter(osw);
//            pw.println(leaderServer);
        zk.reconfig(null, leaderServerId, null, -1, new Stat());
//            pw.close();

        System.exit(10);
//        }
    }

    public static void addRemovedLeader(String server) throws IOException, KeeperException, InterruptedException {
        ZooKeeper zk = new ZooKeeper(server, 30000, new MyWatcher());
        File input = new File("/tmp/removedServer.id");
        if (input.exists()) {
            FileInputStream fis = new FileInputStream(input);
            InputStreamReader isr = new InputStreamReader(fis);
            BufferedReader br = new BufferedReader(isr);
            zk.reconfig(br.readLine(), null, null, -1, new Stat());
            br.close();
            input.delete();

            System.exit(20);
        }
    }

    public static void main(String args[]) {
        if (args.length != 2) {
            System.err.println("usage: ReconfigZkCli <server address> <server port>");
            System.exit(1);
        }

        try {
            String mode = getMode(args[0], args[1]);


            if (mode.equals("observer")) {
                removeLeader(args[0], args[1]);
            }

//            if (mode.equals("follower")) {
//                removeLeader(args[0], args[1]);
//             }
//            if (mode.equals("leader")) {
//                removeLeader(args[0], args[1]);
//            }
//            } else if(mode.equals("follower")) {
//                addRemovedLeader(args[0]);
//            }

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
        } catch (SSLContextException e) {
            System.err.println("SSLContextException: " + e);
            System.exit(1);
        }
    }
}

import org.apache.zookeeper.*;
import org.apache.zookeeper.ZooDefs.*;

public class MyZkCli {

    public static class MyWatcher implements Watcher {
        public void process(WatchedEvent event) {
	    System.out.printf("MyZkCli: Watcher processing event %s (state)\n", event.toString(), event.getState().toString());
        }
    }

    public static void main(String args[]) {
	ZooKeeper zk = null;
        try {
            zk = new ZooKeeper("localhost:2181", 10000, new MyWatcher());

            String data = "fubar";
            for (int i = 0; i < 3; i++) {
		String name = "/test-";
		System.out.printf("MyZkCli: Creating %s (%s)\n", name, data);
		String created = zk.create(name, data.getBytes(), ZooDefs.Ids.OPEN_ACL_UNSAFE, CreateMode.PERSISTENT_SEQUENTIAL);
		System.out.printf("MyZkCli: Created %s\n", created);
            }
        } catch (Exception e) {
	    throw new RuntimeException(e);
        } finally {
	    if ( zk != null ) {
	    	try{
		    System.out.printf("MyZkCli: Closing\n");
	     	    zk.close();
		    System.out.printf("MyZkCli: Closed\n");
	     	} catch(InterruptedException ie){
	     	    throw new RuntimeException(ie);
	     	}
	    }
	}
    }
}

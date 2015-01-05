import HelloApp.*;
import org.omg.CosNaming.*;
import org.omg.CosNaming.NamingContextPackage.*;
import org.omg.CORBA.*;
import org.omg.PortableServer.*;
import org.omg.PortableServer.POA;

import java.util.Properties;

public class HelloImpl extends HelloPOA {
    private ORB orb;

    public void setORB(ORB orb_val) {
	orb = orb_val; 
    }
    
    // implement sayHello() method
    public String sayHello() {
	return "\nHello world !!\n";
    }
    
    // implement shutdown() method
    public void shutdown() {
	orb.shutdown(false);
    }
}


import HelloApp.*;
import org.omg.CosNaming.*;
import org.omg.CosNaming.NamingContextPackage.*;
import org.omg.CORBA.*;

public class HelloClient
{
    static Hello helloImpl;

    public static void main(String args[])
    {
	try{
	    // create and initialize the ORB
	    ORB orb = ORB.init(args, null);

	    // get the root naming context
	    org.omg.CORBA.Object objRef = 
		orb.resolve_initial_references("NameService");
	    // Use NamingContextExt instead of NamingContext. This is 
	    // part of the Interoperable naming Service.  
	    NamingContextExt ncRef = NamingContextExtHelper.narrow(objRef);
 
	    // resolve the Object Reference in Naming
	    String name = "Hello";
	    helloImpl = HelloHelper.narrow(ncRef.resolve_str(name));

	    System.out.println("Obtained a handle on server object: " + helloImpl);
	    System.out.println(helloImpl.sayHello());
	    helloImpl.shutdown();

        } catch (Exception e) {
	    System.out.println("ERROR : " + e) ;
	    e.printStackTrace(System.out);
	}
    }

}


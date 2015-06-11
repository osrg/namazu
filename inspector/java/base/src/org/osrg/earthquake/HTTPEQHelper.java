
import org.osrg.earthquake.*;
import java.util.*;
import org.jboss.byteman.rule.*;
import org.jboss.byteman.rule.helper.*;

public class HTTPEQHelper extends EQHelper
{
    static org.osrg.earthquake.Inspector inspector;
    static {
    	inspector = new org.osrg.earthquake.HTTPInspector();
    };

    HTTPEQHelper(Rule rule) {
	super(rule);
    }
}



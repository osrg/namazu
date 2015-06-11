
import org.osrg.earthquake.*;
import java.util.*;
import org.jboss.byteman.rule.*;
import org.jboss.byteman.rule.helper.*;

public class PBEQHelper extends EQHelper
{
    static {
    	inspector = new org.osrg.earthquake.PBInspector();
    };

    PBEQHelper(Rule rule) {
	super(rule);
    }
}

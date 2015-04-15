import java.util.Scanner;

class Hello {
    private int n;
    public int getN ( ) {
	return this.n;
    }
    public void setN ( int n ) {
	this.n = n;
    }
    
    Hello ( int n ) {
	setN ( n );
    }
    
    private void sayHello ( ) {
	System.out.println ( "hello(): ENTER" );
	int m = n + 1;
	System.out.printf ( "hello(): hello n=%d, m=%d\n", n, m );
	System.out.println ( "hello(): LEAVE" );
    }

    private static int getKeyboardInput ( Scanner keyboard, String prompt ) {
	for ( ; ; ) {
	    System.out.print ( prompt );
	    try {
		String s = keyboard.nextLine ( );
		return Integer.parseInt ( s );
	    } catch ( NumberFormatException nfe ) {
		System.out.println ( "not an integer" );
	    }
	}
    }

    public static void main ( String[] args ) {
	System.out.println ( "main(): ENTER" );
	int i = 42;
	Scanner keyboard = new Scanner ( System.in );
	for ( ; ; ) {
	    Hello h = new Hello ( i );
	    h.sayHello ( );
	    i = getKeyboardInput ( keyboard, "Input an integer: " );
	}
    }
}

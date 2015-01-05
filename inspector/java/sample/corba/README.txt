
This example is based on: http://docs.oracle.com/javase/8/docs/technotes/guides/idl/jidlExample.html

HowTo:
 $ javac HelloApp/*.java *.java
 $ orbd -ORBInitialPort 1050&
 $ java -classpath "/usr/share/java/aspectjrt.jar:./" HelloServer -ORBInitialPort 1050 -ORBInitialHost localhost&
 $ java -classpath "/usr/share/java/aspectjrt.jar:./" HelloClient -ORBInitialPort 1050 -ORBInitialHost localhost


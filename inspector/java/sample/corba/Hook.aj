package HelloApp;

aspect Hook {
       pointcut calling():
       call (String HelloApp.Hello.sayHello());

       before(): calling() {
         System.out.println("aspect method (before)");
       }

       after(): calling() {
         System.out.println("aspect method (after)");
       }
}

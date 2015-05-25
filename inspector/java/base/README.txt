Please run `mvn package` in this directory, then you will get target/earthquake-inspector.jar.
The jar file includes the contents of byteman so that you don't have to concern the class path.


```
export EQ_ENV_PROCESS_ID=foobar
java -javaagent:earthquake-inspector.jar=script:foobar.btm FooBar
```

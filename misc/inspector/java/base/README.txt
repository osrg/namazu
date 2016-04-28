Please run `mvn package` in this directory, then you will get target/nmz-inspector.jar.
The jar file includes the contents of byteman so that you don't have to concern the class path.


```
export NMZ_ENV_PROCESS_ID=foobar
java -javaagent:nmz-inspector.jar=script:foobar.btm FooBar
```

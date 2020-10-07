drinit
=============================================

drinit (`docter init`) is a simple init and process supervisor for exclusive use in Docker contaniers.

What can drinit do?
---------

- drinit execs a single child process, and supervises the process tree lifecycle inside a Docker container.
- drinit runs as PID 1.
- drinit traps and forward signals, and optionally can execute scripts based on trapped signals.
- drinit integrates with Docker HEALTHCHECK.
- drinit reaps zombie processes.
- drinit does not require any changes to your application.


Why not just use the default Docker init ie. krallin/tiny ?
---------

drinit was designed to overcome the caveats of running cloud native software in envrionments like AWS Fargate:

- It is very similar to krallin/tiny, with some of the extra goodness of a process supervisor.
- It works seemlessly with docker HEALTHCHECK. (borrowing the concept of a k8s liveness probe)
- It provides the ability to start, stop and restart applications while the container is running.
- It allows for signal handling workflows using IPC inside the container.

Usage
-----
*NOTE: drinit only works in linux containers

1. Add drinit to your container, and make it executable. 
2. Invoke drinit and pass your program an argument.

In your Docker file, add an ENTRYPOINT and use "Exec form".
```dockerfile
    # Add drinit
    ENV DRINIT_VERSION v0.1.0
    ADD https://github.com/streamz/drinit/releases/download/${DRINIT_VERSION}/drinit .
    ADD https://github.com/streamz/drinit/releases/download/${DRINIT_VERSION}/drinitctl .
    RUN chmod +x drinit
    RUN chmod +x drinitctl
    ENTRYPOINT ["drinit", "--"]  

    # Run your program as a CMD
    CMD ["/your/program", "-and", "-its", "arguments"]
```
The following example traps SIGTERM signal and runs a script:
```dockerfile
     # Add drinit
    ENV DRINIT_VERSION v0.1.0
    ADD https://github.com/streamz/drinit/releases/download/${DRINIT_VERSION}/drinit .
    ADD https://github.com/streamz/drinit/releases/download/${DRINIT_VERSION}/drinitctl .    
    RUN chmod +x drinit
    RUN chmod +x drinitctl
    ADD https://example.com/sigterm.sh sigterm.sh
    RUN chmod +x sigterm.sh
    ADD https://example.com/aftercycle.sh aftercycle.sh
    RUN chmod +x aftercycle.sh
    ENTRYPOINT ["drinit", "-t", "SIGTERM", "-r", "./sigterm.sh", "--"]  

    # Run your program as a CMD
    CMD ["/your/program", "-and", "-its", "arguments"]
    HEALTHCHECK --interval=5s --timeout=3s \
        CMD curl --fail http://localhost:8080/health || ./bounce.sh
 ```

__Features and Options__
-------

## Health Checking ##

drinit leverages the docker [HEALTHCHECK](https://docs.docker.com/engine/reference/builder/#healthcheck]) directive.

You can leverage drinitctl directly or via shell script to perform health checks and actions based on return codes. (0 - healthy, 1 - unhealthy)

## Auto Reaping ##

By default, drinit must run as PID 1 so that it can reap zombies. Any command run by drinit is a child of drinit. The autoreaping feature ensures that any command that is executed does not live as a zombie process in your container./


## Signal Handling ##

drinit can can be configured to trap and execute scripts based on signals it receives. by default, drinit forwards all signals to the supervised process.
<br><br>

How it works
===

After spawning your process, drinit will listen for signals and forward those to the supervised process as long as the signal is not being trapped by the -t switch. Trapped signals will NOT be forwarded to child processes. However, the -t option allows for a -r that can run a script. If a signal is trapped the script run by -r will receive the *nix signum passed as $1. This allows shell scripts to perform custom workflows, and signals can be forwarded by using drinitctl within your script. ex: 

```dockerfile
ENTRYPOINT ["drinit", "-t", "SIGTERM", "-r", "./mysigterm.sh", "--"]  
```

mysigterm.sh
--
``` sh
#!/bin/sh

if [ $1 == 15 ] # sigterm
then
    ./drinitctl -c3 -r dosomething.sh
fi
```

In the example above, when the drinit traps a SIGTERM, it will invoke mysigterm.sh, passing the signum as $1. The script will then use the supervisor control application (drinitctl), to instruct drinit to cycle the application and then run the "dosomething.sh" script. drinit will reap zombie processes that were created within your container by the shell.


Authors
=======

Maintainer:

  + bytecodenerd@gmail.com



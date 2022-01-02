#Kube Drift
## A controller for detecting drift(s) in a Kubernetes clusters
### What is it a drift?
- A drift is a situation where an object in a cluster has changed in a way that is not expected - desired state is not met.
- The controller is designed to detect changes and calculate the drift - difference between the desired state and the current state.
- The focus of this controller is to detect and prevent such a situation.


## Build and Deployment

```bash
 make docker-build docker-push IMG="hugomatus/kube-drift:v1alpha1
```

```bash
make deploy IMG="hugomatus/kube-drift:v1alpha1"
```

```bash
kubectl expose deployment kube-drift-controller-manager -n kube-drift-system --type=NodePort --name=kube-drift --port=8001 --target-port=8001```
```

## References

[Metrics](https://github.com/google/cadvisor/blob/master/docs/storage/prometheus.md)

## Drift API

![demo](assets/kube-drift-api.gif)

## Key output metrics :

### CPU:
#### container_cpu_user_seconds_total: 
    Cumulative “user” CPU time consumed in seconds
#### container_cpu_system_seconds_total:
    Cumulative “system” CPU time consumed in seconds
#### container_cpu_usage_seconds_total: 
    Cumulative CPU time consumed in seconds (sum of the above)
#### container_cpu_cfs_throttled_seconds_total
    This measures the total amount of time a certain container has been throttled. Generally, container CPU usage can be throttled to prevent a single busy container from essentially choking other containers by taking away all the available CPU resources.
    Throttling is usually a good way to ensure a minimum processing power is available for essential services on all running containers. 
    This metric measures the total time that a container’s CPU usage was throttled, and observing this provides the information one needs to properly reallocate resources to specific containers. 
    This can be done, for example, by adjusting the setting for cpu shares in Docker


### container_cpu_load_average_10s

    This measures the value of container CPU load average over the last 10 seconds. Monitoring CPU usage is vital for ensuring it is being used effectively. 
    It would also give insight into what container processes are compute intensive, and as such, help advise future CPU allocation.


### Memory:
#### container_memory_cache: 
    Number of bytes of page cache memory
#### container_memory_swap: 
    Container swap usage in bytes
#### container_memory_usage_bytes: 
    Current memory usage in bytes, including all memory regardless of when it was accessed
    container_memory_max_usage_bytes: Maximum memory usage in byte
### container_memory_failcnt
    This measures the number of times a container’s memory usage limit is hit. It is good practice to set container memory usage limits, to prevent memory intensive tasks from essentially starving other containers on the same server by using all the available memory.
    This way, each container has a max amount of memory they can use, and tracking how many times a container hits its memory usage limit would help a user understand if the container memory limits need to be increased, or if debugging needs to be done in order to find the reason for the high memory consumptions.‍


### Disk:
#### container_fs_io_time_seconds_total: 
    Count of seconds spent doing I/Os
    This measures the cumulative count of seconds spent doing I/Os. It can be used as a baseline to judge the speed of the processes running on your container, and help advise future optimization efforts
#### container_fs_io_time_weighted_seconds_total: 
    Cumulative weighted I/O time in seconds
#### container_fs_writes_bytes_total: 
    Cumulative count of bytes written
#### container_fs_reads_bytes_total: 
    Cumulative count of bytes read

### Network:
##### container_network_receive_bytes_total: 
    Cumulative count of bytes received
#### container_network_receive_errors_total: 
    Cumulative count of errors encountered while receiving
#### container_network_transmit_bytes_total: 
    Cumulative count of bytes transmitted
#### container_network_transmit_errors_total: 
    Cumulative count of errors encountered while transmitting

### Tasks and Processes:

#### container_processes
    This metric keeps track of the number of processes currently running inside the container. Knowing the exact state of our containers at all times is essential in keeping them up and running. As such, knowing how many processes are currently running in a specific container would provide insight into whether things are functioning normally, or whether there’s something wrong.‍
#### container_tasks_state
    This metric tracks the number of tasks or processes in a given state (sleeping, running, stopped, uninterruptible, or ioawaiting) in a container. At a glance, this information could be essential in providing real-time information on the status or health of the container and its processes.‍
#### container_start_time_seconds
    Although subtle, this metric tracks a container’s start time in seconds, and could either provide an early indication of trouble, or an indication of a healthy container instance.
# process-exporter
Prometheus exporter that mines /proc to report on selected processes.

The premise for this exporter is that sometimes you have apps that are
impractical to instrument directly, either because you don't control the code
or they're written in a language that isn't easy to instrument with Prometheus.
A fair bit of information can be gleaned from /proc, especially for
long-running programs.

For most systems it won't be beneficial to create metrics for every process by
name: there are just too many of them and most don't do enough to merit it.
Various command-line options are provided to control how processes are grouped
and the groups are named.  Run "process-exporter -man" to see a help page
giving details.

Metrics available currently include CPU usage, bytes written and read, and
number of processes in each group.  

Bytes read and written come from /proc/[pid]/io in recent enough kernels.
These correspond to the fields `read_bytes` and `write_bytes` respectively.
These IO stats come with plenty of caveats, see either the Linux kernel 
documentation or man 5 proc.

CPU usage comes from /proc/[pid]/stat fields utime (user time) and stime (system
time.)  It has been translated into fractional seconds of CPU consumed.  Since
it is a counter, using rate() will tell you how many fractional cores were running
code from this process during the interval given.

An example Grafana dashboard to view the metrics is available at https://grafana.net/dashboards/249

## Instrumentation cost

process-exporter will consume CPU in proportion to the number of processes in
the system and the rate at which new ones are created.  The most expensive
parts - applying regexps and executing templates - are only applied once per
process seen.  If you have mostly long-running processes process-exporter
should be lightweight: each time a scrape occurs, parsing of /proc/$pid/stat
and /proc/$pid/cmdline for every process being monitored and adding a few
numbers.

## Config

To select and group the processes to monitor, either provide command-line
arguments or use a YAML configuration file. 

To avoid confusion with the cmdline YAML element, we'll refer to the
null-delimited contents of `/proc/<pid>/cmdline` as the array `argv[]`.

Each item in `process_names` gives a recipe for identifying and naming
processes.  The optional `name` tag defines a template to use to name
matching processes; if not specified, `name` defaults to `{{.ExeBase}}`.

Template variables available:
- `{{.ExeBase}}` contains the basename of the executable
- `{{.ExeFull}}` contains the fully qualified path of the executable
- `{{.Matches}}` map contains all the matches resulting from applying cmdline regexps

Each item in `process_names` must contain one or more selectors (`comm`, `exe`
or `cmdline`); if more than one selector is present, they must all match.  Each
selector is a list of strings to match against a process's `comm`, `argv[0]`,
or in the case of `cmdline`, a regexp to apply to the command line.  

For `comm` and `exe`, the list of strings is an OR, meaning any process
matching any of the strings will be added to the item's group.  

For `cmdline`, the list of regexes is an AND, meaning they all must match.  Any
capturing groups in a regexp must use the `?P<name>` option to assign a name to
the capture, which is used to populate `.Matches`.

A process may only belong to one group: even if multiple items would match, the
first one listed in the file wins.

Other performance tips: give an exe or comm clause in addition to any cmdline
clause, so you avoid executing the regexp when the executable name doesn't
match.

```

process_names:
  # comm is the second field of /proc/<pid>/stat minus parens.
  # It is the base executable name, truncated at 15 chars.  
  # It cannot be modified by the program, unlike exe.
  - comm:
    - bash
    
  # exe is argv[0]. If no slashes, only basename of argv[0] need match.
  # If exe contains slashes, argv[0] must match exactly.
  - exe: 
    - postgres
    - /usr/local/bin/prometheus

  # cmdline is a list of regexps applied to argv.
  # Each must match, and any captures are added to the .Matches map.
  - name: "{{.ExeFull}}:{{.Matches.Cfgfile}}"
    exe: 
    - /usr/local/bin/process-exporter
    cmdline: 
    - -config.path\\s+(?P<Cfgfile>\\S+)
    

```

Here's the config I use on my home machine:

```

process_names:
  - comm: 
    - chromium-browse
    - bash
    - prometheus
    - gvim
  - exe: 
    - /sbin/upstart
    cmdline: 
    - --user
    name: upstart:-user

```

## Docker

A docker image can be created with

```
make docker
```

Then run the docker, e.g.

```
docker run --privileged --name pexporter -d -v /proc:/host/proc -p 127.0.0.1:9256:9256 process-exporter:master -procfs /host/proc -procnames chromium-browse,bash,prometheus,gvim,upstart:-user -namemapping "upstart,(-user)"
```

This will expose metrics on http://localhost:9256/metrics.  Leave off the
`127.0.0.1:` to publish on all interfaces.  Leave off the --priviliged and
add the --user docker run argument if you only need to monitor processes
belonging to a single user.

## History

An earlier version of this exporter had options to enable auto-discovery of
which processes were consuming resources.  This functionality has been removed.
These options were based on a percentage of resource usage, e.g. if an
untracked process consumed X% of CPU during a scrape, start tracking processes
with that name.  However during any given scrape it's likely that most
processes are idle, so we could add a process that consumes minimal resources
but which happened to be active during the interval preceding the current
scrape.  Over time this means that a great many processes wind up being
scraped, which becomes unmanageable to visualize.  This could be mitigated by
looking at resource usage over longer intervals, but ultimately I didn't feel
this feature was important enough to invest more time in at this point.  It may
re-appear at some point in the future, but no promises.

Another lost feature: the "other" group was used to count usage by non-tracked
procs.  This was useful to get an idea of what wasn't being monitored.  But it
comes at a high cost: if you know what processes you care about, you're wasting
a lot of CPU to compute the usage of everything else that you don't care about.
The new approach is to minimize resources expended on non-tracked processes and
to require the user to whitelist the processes to track.  

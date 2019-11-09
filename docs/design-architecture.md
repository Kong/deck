# Design & Architecture

## Underlying architecture

### Reverse sync

One of the most important features of decK is reverse-sync, whereby decK can
detect entities that are present in Kong's database but are not part of the
state file.
This feature increases the complexity of the project as the code needs to
perform a sync in both directions, from the state file to Kong and from Kong
to the state file.

### Algorithm

#### Export and Reset

An export or reset of entities is fairly easy to implement.
decK loads all the entities from Kong into memory and then serializes
it into a YAML or JSON file. For reset, it instead performs `DELETE` queries
on all the entities.

#### Diff and Sync

The `diff` of configuration is performed using the following algorithm:

1. Read the configuration from Kong and store it in a SQL-like in-memory
   database.
1. Read the state file from disk, and match the `ID`s of entity with their
   respective counterparts in the in-memory state, if they are present.
1. Now, for entity of each type we perform the following:
   1. *Create*: if the entity is not present in Kong, create the entity.
   1. *Update*: if the entity is present in Kong, check for equality. If not
      equal, then update it in Kong. These two steps are referred to as
      "forward sync".
   1. *Delete*: Go through each entity in Kong (from the in-memory database),
      and check if it is present in the state file, if yes, don't do anything.
      If no, then delete the entity from Kong's database as well.

Certain filters like `select-tag` or Kong Enterprise workspace might be applied
to the above algorithm based on the inputs given to decK.

### Operational outlook

Based on the above algorithm, one can see how decK can require a large amount
of memory and network I/O. While this is true, a few optimizations have
been incorporated to ensure good performance:
- For network operations, decK minimizes the API calls it has to make to Kong
  to read the state. It uses list endpoints in Kong with a large page size
  (`1000`) for efficiency.
- decK parallelizes various Create/Update/Delete operations where it can. So,
  if decK and Kong or Kong and Kong's database are present far apart in terms
  of network latency, parallel operations help speed up operations.
  With smaller installations, this optimization might not be measurable.
- decK's memory footprint can be high if the configuration for Kong is huge.
  This is usually not a concern as decK's process is short-lived. For very
  large installation, it is recommended to configure a sub-set of
  the large configuration at one time using a technique referred to as
  [distributed configuration](guides/distributed-configuration.md).
  There are avenues to further reduce the memory requirements of decK,
  although, we don't know by how much. decK's code is written with focus on
  correctness over performance.

## Choice of language

decK is written in Go because:
- Go provides good concurrency primitives which helps ensuring high-performance
  for decK.
- Go's compiler spits out a static compiled binary, meaning no other dependency
  need to be installed on the system. This gives a very good end-user experience
  as installing downloading and copying a single binary is easy and fast.
- decK original goal was much larger than what it is today. If we decide to
  pursue larger goals(think a control-plane for Kong) in future,
  Go is probably the best language available to write that type of software.
- the original author was familiar with Go :)


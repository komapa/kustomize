// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by "mdtogo"; DO NOT EDIT.
package api

var ConfigFnLong = `# Configuration Functions API Semantics

  Configuration Functions are functions packaged as executables in containers which enable
  **shift-left practices**.  They configure applications and infrastructure through
  Kubernetes style Resource Configuration, but run locally pre-commit.

  Configuration functions enable shift-left practices (client-side) through:

  - Pre-commit / delivery validation and linting of configuration
    - e.g. Fail if any containers don't have PodSecurityPolicy or CPU / Memory limits
  - Implementation of abstractions as client actuated APIs (e.g. templating)
    - e.g. Create a client-side *"CRD"* for generating configuration checked into git
  - Aspect Orient configuration / Injection of cross-cutting configuration
    - e.g. T-Shirt size containers by annotating Resources with ` + "`" + `small` + "`" + `, ` + "`" + `medium` + "`" + `, ` + "`" + `large` + "`" + `
      and inject the cpu and memory resources into containers accordingly.
    - e.g. Inject ` + "`" + `init` + "`" + ` and ` + "`" + `side-car` + "`" + ` containers into Resources based off of Resource
      Type, annotations, etc.

  Performing these on the client rather than the server enables:

  - Configuration to be reviewed prior to being sent to the API server
  - Configuration to be validated as part of the CD pipeline
  - Configuration for Resources to validated holistically rather than individually
    per-Resource -- e.g. ensure the ` + "`" + `Service.selector` + "`" + ` and ` + "`" + `Deployment.spec.template` + "`" + ` labels
    match.
    - MutatingWebHooks are scoped to a single Resource instance at a time.
  - Low-level tweaks to the output of high-level abstractions -- e.g. add an ` + "`" + `init container` + "`" + `
    to a client *"CRD"* Resource after it was generated.
  - Composition and layering of multiple functions together
    - Compose generation, injection, validation together

  Configuration Functions are implemented as executable programs published in containers which:

  - Accept as input (stdin):
    - A list of Resource Configuration
    - A Function Configuration (to configure the function itself)
  - Emit as output (stdout + exit):
    - A list of Resource Configuration
    - An exit code for success / failure

### Function Specification

  - Functions **SHOULD** be published as container images containing a ` + "`" + `CMD` + "`" + ` invoking an executable.
  - Functions **MUST** accept input on STDIN a ` + "`" + `ResourceList` + "`" + ` containing the Resources and
    ` + "`" + `functionConfig` + "`" + `.
  - Functions **MUST** emit output on STDOUT a ` + "`" + `ResourceList` + "`" + ` containing the modified
    Resources.
  - Functions **MUST** exit non-0 on failure, and exit 0 on success.
  - Functions **MAY** emit output on STDERR with error messaging.
  - Functions performing validation **SHOULD** exit failure and emit error messaging
    on a validation failure.
  - Functions generating Resources **SHOULD** retain non-conflicting changes on the
    generated Resources -- e.g. 1. the function generates a Deployment, but doesn't
    specify ` + "`" + `cpu` + "`" + `, 2. the user sets the ` + "`" + `cpu` + "`" + ` on the generated Resource, 3. the
    function should keep the ` + "`" + `cpu` + "`" + ` when regenerating the Resource a second time.
  - Functions **SHOULD** be usable outside ` + "`" + `kustomize config run` + "`" + ` -- e.g. though pipeline
    mechanisms such as Tekton.

#### Input Format

  Functions must accept on STDIN:

  ` + "`" + `ResourceList` + "`" + `:
  - contains ` + "`" + `items` + "`" + ` field, same as ` + "`" + `List.items` + "`" + `
  - contains ` + "`" + `functionConfig` + "`" + ` field -- a single item with the configuration for the function itself

  Example ` + "`" + `ResourceList` + "`" + ` Input:

    apiVersion: config.kubernetes.io/v1alpha1
    kind: ResourceList
    functionConfig:
      apiVersion: example.com/v1beta1
      kind: Nginx
      metadata:
        name: my-instance
        annotations:
          config.kubernetes.io/local-config: "true"
      spec:
        replicas: 5
    items:
    -  apiVersion: apps/v1
       kind: Deployment
       metadata:
         name: my-instance
       spec:
         replicas: 3
         ...
    - apiVersion: v1
      kind: Service
      metadata:
        name: my-instance
      spec:
        ...

#### Output Format

  Functions must emit on STDOUT:

  ` + "`" + `ResourceList` + "`" + `:
  - contains ` + "`" + `items` + "`" + ` field, same as ` + "`" + `List.items` + "`" + `

  Example ` + "`" + `ResourceList` + "`" + ` Output:

    apiVersion: config.kubernetes.io/v1alpha1
    kind: ResourceList
    items:
    -  apiVersion: apps/v1
       kind: Deployment
       metadata:
         name: my-instance
       spec:
         replicas: 5
         ...
    - apiVersion: v1
      kind: Service
      metadata:
        name: my-instance
      spec:
        ...

#### Container Environment

  When run by ` + "`" + `kustomize config run` + "`" + `, functions are run in containers with the
  following environment:

  - Network: ` + "`" + `none` + "`" + `
  - User: ` + "`" + `nobody` + "`" + `
  - Security Options: ` + "`" + `no-new-privileges` + "`" + `
  - Volumes: the volume containing the ` + "`" + `functionConfig` + "`" + ` yaml is mounted under ` + "`" + `/local` + "`" + ` as ` + "`" + `ro` + "`" + `

### Example Function Implementation

  Following is an example for implementing an nginx abstraction using a config
  function.

#### ` + "`" + `nginx-template.sh` + "`" + `

  ` + "`" + `nginx-template.sh` + "`" + ` is a simple bash script which uses a *heredoc* as a templating solution
  for generating Resources from the functionConfig input fields.

  The script wraps itself using ` + "`" + `config run wrap -- $0` + "`" + ` which will:

  1. Parse the ` + "`" + `ResourceList.functionConfig` + "`" + ` (provided to the container stdin) into env vars
  2. Merge the stdout into the original list of Resources
  3. Defaults filenames for newly generated Resources (if they are not set as annotations)
     to ` + "`" + `config/NAME_KIND.yaml` + "`" + `
  4. Format the output

    #!/bin/bash
    # script must run wrapped by ` + "`" + `kustomize config run wrap` + "`" + `
    # for parsing input the functionConfig into env vars
    if [ -z ${WRAPPED} ]; then
      export WRAPPED=true
      config run wrap -- $0
      exit $?
    fi

    cat <<End-of-message
    apiVersion: v1
    kind: Service
    metadata:
      name: ${NAME}
      labels:
        app: nginx
        instance: ${NAME}
    spec:
      ports:
      - port: 80
        targetPort: 80
        name: http
      selector:
        app: nginx
        instance: ${NAME}
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: ${NAME}
      labels:
        app: nginx
        instance: ${NAME}
    spec:
      replicas: ${REPLICAS}
      selector:
        matchLabels:
          app: nginx
          instance: ${NAME}
      template:
        metadata:
          labels:
            app: nginx
            instance: ${NAME}
        spec:
          containers:
          - name: nginx
            image: nginx:1.7.9
            ports:
            - containerPort: 80
    End-of-message

#### ` + "`" + `Dockerfile` + "`" + `

  ` + "`" + `Dockerfile` + "`" + ` installs ` + "`" + `kustomize config` + "`" + ` and copies the script into the container image.

    FROM golang:1.13-stretch
    RUN go get sigs.k8s.io/kustomize/cmd/config
    RUN mv /go/bin/config /usr/bin/config
    COPY nginx-template.sh /usr/bin/nginx-template.sh
    CMD ["nginx-template.sh]

### Example Function Usage

Following is an example of running the ` + "`" + `kustomize config run` + "`" + ` using the preceding API.

#### ` + "`" + `nginx.yaml` + "`" + ` (Input)

  ` + "`" + `dir/nginx.yaml` + "`" + ` contains a reference to the Function.  The contents of ` + "`" + `nginx.yaml` + "`" + `
  are passed to the Function through the ` + "`" + `ResourceList.functionConfig` + "`" + ` field.

    apiVersion: example.com/v1beta1
    kind: Nginx
    metadata:
      name: my-instance
      annotations:
        config.kubernetes.io/local-config: "true"
      configFn:
        container:
          image: gcr.io/example-functions/nginx-template:v1.0.0
    spec:
      replicas: 5

  - ` + "`" + `configFn.container.image` + "`" + `: the image to use for this API
  - ` + "`" + `annotations[config.kubernetes.io/local-config]` + "`" + `: mark this as not a Resource that should
    be applied

#### ` + "`" + `kustomize config run dir/` + "`" + ` (Output)

  ` + "`" + `dir/my-instance_deployment.yaml` + "`" + ` contains the Deployment:

    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: my-instance
      labels:
        app: nginx
        instance: my-instance
    spec:
      replicas: 5
      selector:
        matchLabels:
          app: nginx
          instance: my-instance
      template:
        metadata:
          labels:
            app: nginx
            instance: my-instance
        spec:
          containers:
          - name: nginx
            image: nginx:1.7.9
            ports:
            - containerPort: 80

  ` + "`" + `dir/my-instance_service.yaml` + "`" + ` contains the Service:

    apiVersion: v1
    kind: Service
    metadata:
      name: my-instance
      labels:
        app: nginx
        instance: my-instance
    spec:
      ports:
      - port: 80
        targetPort: 80
        name: http
      selector:
        app: nginx
        instance: my-instance`

var ConfigIoLong = `# Configuration IO API Semantics

Resource Configuration may be read / written from / to sources such as directories,
stdin|out or network. Tools may be composed using pipes such that the tools writing
Resource Configuration may be a different tool from the one that read the configuration.
In order for tools to be composed in this way, while preserving origin information --
such as the original file, index, etc.

Tools **SHOULD** write the following annotations when reading from sources,
and **SHOULD** respect the annotations when writing to sinks.

### ` + "`" + `config.kubernetes.io/path` + "`" + `

Records the slash-delimited, OS-agnostic, relative file path to a Resource.

This annotation **SHOULD** be set when reading Resources from files.
When writing Resources to a directory, the Resource **SHOULD** be written to the corresponding
path relative to that directory.

Example:

	metadata:
	  annotations:
	    config.kubernetes.io/path: "relative/file/path.yaml"

### ` + "`" + `config.kubernetes.io/index` + "`" + `

Records the index of a Resource in file. In a multi-object files YAML file, Resources are separated
by three dashes (` + "`" + `---` + "`" + `), and the index represents the positon of the Resource starting from zero.

This annotation **SHOULD** be set when reading Resources from files.
When writing multiple Resources to the same file, the Resource **SHOULD** be written in the
relative order matching the index.

When this annotation is not specified, it implies a value of ` + "`" + `0` + "`" + `.

Example:

	metadata:
	  annotations:
	    config.kubernetes.io/path: "relative/file/path.yaml"
	    config.kubernetes.io/index: 2

This represents the third Resource in the file.

### ` + "`" + `config.kubernetes.io/local-config` + "`" + `

` + "`" + `config.kubernetes.io/local-config` + "`" + ` declares that the configuration is to local tools
rather than a remote Resource. e.g. The ` + "`" + `Kustomization` + "`" + ` config in a ` + "`" + `kustomization.yaml` + "`" + `
**SHOULD** contain this annotation so that tools know it is not intended to be sent to
the Kubernetes api server.

Example:

	metadata:
	  annotations:
	    config.kubernetes.io/local-config: "true"`

var Merge2Long = `# Merge (2-way)

  2-way merges fields from a source to a destination, overriding the destination fields
  where they differ.

  ### Merge Rules

  Fields are recursively merged using the following rules:

  - scalars
    - if present only in the dest, it keeps its value
    - if present in the src and is non-null, take the src value -- if ` + "`" + `null` + "`" + `, clear it
    - example src: ` + "`" + `5` + "`" + `, dest: ` + "`" + `3` + "`" + ` => result: ` + "`" + `5` + "`" + `

  - non-associative lists -- lists without a merge key
    - if present only in the dest, it keeps its value
    - if present in the src and is non-null, take the src value -- if ` + "`" + `null` + "`" + `, clear it
    - example src: ` + "`" + `[1, 2, 3]` + "`" + `, dest: ` + "`" + `[a, b, c]` + "`" + ` => result: ` + "`" + `[1, 2, 3]` + "`" + `

  - map keys and fields -- paired by the map-key / field-name
    - if present only in the dest, it keeps its value
    - if present only in the src, it is added to the dest
    - if the field is present in both the src and dest, and the src value is
      ` + "`" + `null` + "`" + `, the field is removed from the dest
    - if the field is present in both the src and dest, the value is recursively merged
    - example src: ` + "`" + `{'key1': 'value1', 'key2': 'value2'}` + "`" + `,
      dest: ` + "`" + `{'key2': 'value0', 'key3': 'value3'}` + "`" + `
      => result: ` + "`" + `{'key1': 'value1', 'key2': 'value2', 'key3': 'value3'}` + "`" + `

  - associative list elements -- paired by the associative key
    - if present only in the dest, it keeps its value in the list
    - if present only in the src, it is added to the dest list
    - if the field is present in both the src and dest, the value is recursively merged

  ### Associative Keys

  Associative keys are used to identify "same" elements within 2 different lists, and merge them.
  The following fields are recognized as associative keys:

  [` + "`" + `mountPath` + "`" + `, ` + "`" + `devicePath` + "`" + `, ` + "`" + `ip` + "`" + `, ` + "`" + `type` + "`" + `, ` + "`" + `topologyKey` + "`" + `, ` + "`" + `name` + "`" + `, ` + "`" + `containerPort` + "`" + `]

  Any lists where all of the elements contain associative keys will be merged as associative lists.

  ### Example

  > Source

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 3 # scalar
	  template:
	    spec:
	      containers:  # associative list -- (name)
	      - name: nginx
	        image: nginx:1.7
	        command: ['new_run.sh', 'arg1'] # non-associative list
	      - name: sidecar2
	        image: sidecar2:v1

  > Destination

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 1
	  template:
	    spec:
	      containers:
	      - name: nginx
	        image: nginx:1.6
	        command: ['old_run.sh', 'arg0']
	      - name: sidecar1
	        image: sidecar1:v1

  > Result

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 3 # scalar
	  template:
	    spec:
	      containers:  # associative list -- (name)
	      - name: nginx
	        image: nginx:1.7
	        command: ['new_run.sh', 'arg1'] # non-associative list
	      - name: sidecar1
	        image: sidecar1:v1
	      - name: sidecar2
	        image: sidecar2:v1`

var Merge3Long = `# Merge (3-way)

  3-way merge identifies changes between an original source + updated source and merges the result
  into a destination, overriding the destination fields where they have changed between
  original and updated.

### Resource MergeRules

  - Resources present in the original and deleted from the update are deleted.
  - Resources missing from the original and added in the update are added.
  - Resources present only in the dest are kept without changes.
  - Resources present in both the update and the dest have their fields merged with the destination.

### Field Merge Rules

  Fields are recursively merged using the following rules:

  - scalars
    - if present in either dest or updated and ` + "`" + `null` + "`" + `, clear the value
    - if unchanged between original and updated, keep dest value
    - if changed between original and updated (added, deleted, changed), take the updated value

  - non-associative lists -- lists without a merge key
    - if present in either dest or updated and ` + "`" + `null` + "`" + `, clear the value
    - if unchanged between original and updated, keep dest value
    - if changed between original and updated (added, deleted, changed), take the updated value

  - map keys and fields -- paired by the map-key / field-name
    - if present in either dest or updated and ` + "`" + `null` + "`" + `, clear the value
    - if present only in the dest, it keeps its value
    - if not-present in the dest, add the delta between original-updated as a field
    - otherwise recursively merge the value between original, updated, dest

  - associative list elements -- paired by the associative key
    - if present only in the dest, it keeps its value
    - if not-present in the dest, add the delta between original-updated as a field
    - otherwise recursively merge the value between original, updated, dest

### Associative Keys

  Associative keys are used to identify "same" elements within 2 different lists, and merge them.
  The following fields are recognized as associative keys:

  [` + "`" + `mountPath` + "`" + `, ` + "`" + `devicePath` + "`" + `, ` + "`" + `ip` + "`" + `, ` + "`" + `type` + "`" + `, ` + "`" + `topologyKey` + "`" + `, ` + "`" + `name` + "`" + `, ` + "`" + `containerPort` + "`" + `]

  Any lists where all of the elements contain associative keys will be merged as associative lists.`

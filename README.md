# jobctl

**A simple command to run workflows (DAGs)**

jobctl is a single command that generates and executes a [DAG (Directed acyclic graph)](https://en.wikipedia.org/wiki/Directed_acyclic_graph) from a simple YAML definition. jobctl also comes with a convenient web UI & REST api interface. It aims to be one of the easiest option to manage DAGs executed by cron.

## Contents
- [jobctl](#jobctl)
  - [Contents](#contents)
  - [Motivation](#motivation)
  - [Why not alternatives, like Airflow?](#why-not-alternatives-like-airflow)
  - [Installation](#installation)
  - [Features](#features)
  - [Use cases](#use-cases)
  - [User interface](#user-interface)
  - [Architecture](#architecture)
  - [Getting started](#getting-started)
    - [Installation](#installation-1)
    - [Usage](#usage)
  - [Configuration](#configuration)
    - [Environment variables](#environment-variables)
    - [Web UI configuration](#web-ui-configuration)
    - [Global configuration](#global-configuration)
  - [Job configuration](#job-configuration)
    - [Simple configuration](#simple-configuration)
    - [All configuration](#all-configuration)
  - [Examples](#examples)
  - [FAQ](#faq)
    - [How to contribute?](#how-to-contribute)
  - [Todo](#todo)
  - [License](#license)

## Motivation
Currently, my environment has **many problems**. Hundreds of complex cron jobs are registered on huge servers and it is impossible to keep track of the dependencies between them. If one job fails, I don't know which job to re-run. I also have to SSH into the server to see the logs and manually run the shell scripts one by one.

So I needed a tool that can explicitly visualize and manage the dependencies of the pipeline.

***How nice it would be to be able to visually see the job dependencies, execution status, and logs of each job in a web browser, and to be able to rerun or stop a series of jobs with just a mouse click!***

## Why not alternatives, like Airflow?
I considered many potential tools such as Airflow, Rundeck, Luigi, DigDag, JobScheduler, etc.

But unfortunately, they were not suitable for my existing environment. Because they required a DBMS(Database Management System) installation, relatively high learning curves, and more operational overheads. We only have a small group of engineers in our office and use a less common DBMS.

Finally, I decided to build my own tool that would not require any DBMS server, any daemon process, or any additional operational burden and is easy to use.

## Installation
Download the binary from [Releases page](https://github.com/jobctl/jobctl/releases) and place it on your system.

## Features

- Simple command interface (See [Usage](#usage))
- Simple configuration YAML format (See [Simple example](#simple-example))
- Simple architecture (no DBMS or agent process is required)
- Web UI to visualize, manage jobs and watch logs
- Parameterization
- Conditions
- Automatic retry
- Cancellation
- Retry
- Prallelism limits
- Environment variables
- Repeat
- Basic Authentication
- E-mail notifications
- REST api interface
- onExit / onSuccess / onFailure / onCancel handlers
- Automatic data cleaning

## Use cases
- ETL Pipeline
- Batches
- Machine Learning
- Data Processing
- Automation

## User interface

- **JOBs**: Overview of all JOBs in your environment.

  ![JOBs](https://user-images.githubusercontent.com/1475839/164859814-98afc587-ff86-4ebd-97b8-7d32f86a9ad9.png)

- **Detail**: Current status of the job.

  ![Detail](https://user-images.githubusercontent.com/1475839/164857046-b620a8a0-f5f5-4551-a651-66c8ec38f820.png)

- **Timeline**: Timeline of each steps in the pipeline.

  ![Timeline](https://user-images.githubusercontent.com/1475839/164860845-98595a3f-4579-4c15-9d6b-1942b4561900.png)

- **History**: History of the execution of the pipeline.

  ![History](https://user-images.githubusercontent.com/1475839/164849560-ab5be8d0-378e-46eb-a4d4-c6a8ff3d6af9.png)

## Architecture

- uses plain JSON files as history database, and unix sockets to communicate with running processes.
  ![jobctl Architecture](https://user-images.githubusercontent.com/1475839/164869015-769bfe1d-ad38-4aca-836b-bf3ffe0665df.png)

## Getting started
### Installation

Place a `jobctl` executable somewhere on your system.

### Usage

- `jobctl start [--params=<params>] <job file>` - run a job
- `jobctl status <job file>` - display the current status of the job
- `jobctl retry --req=<request-id> <job file>` - retry the failed/canceled job
- `jobctl stop <job file>` - cancel a job
- `jobctl dry [--params=<params>] <job file>` - dry-run a job
- `jobctl server` - start a web server for web UI

## Configuration

### Environment variables
- `JOBCTL__DATA` - path to directory for internal use by jobctl (default : `~/.jobctl/data`)
- `JOBCTL__LOGS` - path to directory for logging (default : `~/.jobctl/logs`)

### Web UI configuration

Plase create `~/.jobctl/admin.yaml`.

```yaml
# required
host: <hostname for web UI address>
port: <port number for web UI address>
jobs: <the location of job configuration files>
command: <absolute path to the jobctl exectable>

# optional
isBasicAuth: <true|false>
basicAuthUsername: <username for basic auth of web UI>
basicAuthPassword: <password for basic auth of web UI>
```

### Global configuration

Plase create `~/.jobctl/config.yaml`. All settings can be overridden by individual job configurations.

```yaml
logDir: <path-to-write-log>   # log directory to write standard output from the job steps
histRetentionDays: 3 # job history retention days (not for log files)

# E-mail server config (optional)
smtp:
  host: <smtp server host>
  port: <stmp server port>
errorMail:
  from: <from address to send error mail>
  to: <to address to send error mail>
  prefix: <prefix of mail subject for error mail>
infoMail:
  from: <from address to send notification mail>
  to: <to address to send notification mail>
  prefix: <prefix of mail subject for notification mail>
```

## Job configuration

### Simple configuration

```yaml
name: simple configuration
steps:
  - name: step 1
    command: python some_batch_1.py
    dir: ${HOME}/jobs/                  # working directory for the job (optional)
  - name: step 2
    command: python some_batch_2.py
    dir: ${HOME}/jobs/
    depends:
      - step 1
```

### All configuration

```yaml
name: all configuration
description: run python jobs

# Define environment variables
env:
  LOG_DIR: ${HOME}/jobs/logs
  PATH: /usr/local/bin:${PATH}
  
logDir: ${LOG_DIR}   # log directory to write standard output from the job steps
histRetentionDays: 3 # job history retention days (not for log files)
delaySec: 1          # interval seconds between job steps
maxActiveRuns: 1     # max parallel number of running step

# Define parameters
params: param1 param2 # they can be referenced by each steps as $1, $2 and so on.

# Define preconditions for whether or not the job is allowed to run
preconditions:
  - condition: "`printf 1`" # This condition will be evaluated at each execution of the job
    expected: "1"           # If the evaluation result do not match, the job is canceled

# Mail notification configs
mailOnError: true    # send a mail when a job failed
mailOnFinish: true   # send a mail when a job finished

# Handler on Success, Failure, Cancel, Exit
handlerOn:
  success:
    command: "echo succeed"
  failure:
    command: "echo failed"
  cancel:
    command: "echo canceled"
  exit:
    command: "echo finished"

# Job steps
steps:
  - name: step 1
    description: step 1 description
    dir: ${HOME}/jobs
    command: python some_batch_1.py $1
    mailOnError: false # do not send mail on error
    continueOn:
      failed: true     # continue to the next step regardless the error of this job
      skipped: true    # continue to the next step regardless the evaluation result of preconditions
    retryPolicy:
      limit: 2         # retry up to 2 times when the step failed
    # Define preconditions for whether or not the step is allowed to run
    preconditions:
      - condition: "`printf 1`"
        expected: "1"
  - name: step 2
    description: step 2 description
    dir: ${HOME}/jobs
    command: python some_batch_2.py $1
    depends:
      - step 1
```

The global config file `~/.jobctl/config.yaml` is useful to gather common settings such as mail-server configs or log directory.

## Examples

To check all examples, visit [this page](https://github.com/jobctl/jobctl/tree/main/examples).

-  Sample 1

  ![sample_1](https://user-images.githubusercontent.com/1475839/164965036-fede5675-cba0-410b-b371-22deec55b9e8.png)
```yaml
name: example job
steps:
  - name: "1"
    command: echo hello world
  - name: "2"
    command: sleep 10
    depends:
      - "1"
  - name: "3"
    command: echo done!
    depends:
      - "2"
```

-  Sample 2

  ![sample_2](https://user-images.githubusercontent.com/1475839/164965143-b10a0511-35f3-45fa-9eba-69c6db4614a2.png)
```yaml
name: example job
env:
  LOG_DIR: ${HOME}/logs
logDir: ${LOG_DIR}
params: foo bar
steps:
  - name: "check precondition"
    command: echo start
    preconditions:
      - condition: "`echo $1`"
        expected: foo
  - name: "print foo"
    command: echo $1
    depends:
      - "check precondition"
  - name: "print bar"
    command: echo $2
    depends:
      - "print foo"
  - name: "failure and continue"
    command: "false"
    continueOn:
      failure: true
    depends:
      - "print bar"
  - name: "print done"
    command: echo done!
    depends:
      - "failure and continue"
handlerOn:
  exit:
    command: echo finished!
  success:
    command: echo success!
  failure:
    command: echo failed!
  cancel:
    command: echo canceled!
```

-  Complex example

  ![complex](https://user-images.githubusercontent.com/1475839/164965345-977de1bc-d042-4d3f-bf0e-bb648e534a78.png)
```yaml
name: complex job
steps:
  - name: "Initialize"
    command: "sleep 2"
  - name: "Copy TAB_1"
    description: "Extract data from TAB_1 to TAB_2"
    command: "sleep 2"
    depends:
      - "Initialize"
  - name: "Update TAB_2"
    description: "Update TAB_2"
    command: "sleep 2"
    depends:
      - Copy TAB_1
  - name: Validate TAB_2
    command: "sleep 2"
    depends:
      - "Update TAB_2"
  - name: "Load TAB_3"
    description: "Read data from files"
    command: "sleep 2"
    depends:
      - Initialize
  - name: "Update TAB_3"
    command: "sleep 2"
    depends:
      - "Load TAB_3"
  - name: Merge
    command: "sleep 2"
    depends:
      - Update TAB_3
      - Validate TAB_2
      - Validate File
  - name: "Check File"
    command: "sleep 2"
  - name: "Copy File"
    command: "sleep 2"
    depends:
      - Check File
  - name: "Validate File"
    command: "sleep 2"
    depends:
      - Copy File
  - name: Calc Result
    command: "sleep 2"
    depends:
      - Merge
  - name: "Report"
    command: "sleep 2"
    depends:
      - Calc Result
  - name: Reconcile
    command: "sleep 2"
    depends:
      - Calc Result
  - name: "Cleaning"
    command: "sleep 2"
    depends:
      - Reconcile
```

## FAQ

### How to contribute?

Feel free to contribute in any way you want. Share ideas, submit issues, create pull requests. 
You can start by improving this [README.md](https://github.com/jobctl/jobctl/blob/main/README.md) or suggesting new [features](https://github.com/jobctl/jobctl/issues)
Thank you!

## Todo

- [ ] Documentation for YAML definitions
- [ ] Prettier CLI interface
- [ ] JWT authentication
- [ ] History compaction
- [ ] History sub command
- [ ] Edit YAML on Web UI
- [ ] Pause & Resume pipeline
- [ ] Docker container

## License
This project is licensed under the GNU GPLv3 - see the [LICENSE.md](LICENSE.md) file for details
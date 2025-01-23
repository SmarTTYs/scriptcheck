# Scriptcheck

Simple CLI to extract or shellcheck scripts inlined in pipeline yaml files, supporting
multiple ci/cd formats.

Currently only support for gitlab is implemented as an MVP.

NOTE: This project is currently only provided as an MVP for testing purposes
at my workspace. No further support is provided.

## Supported Formats
- Gitlab CI/CD

## Gitlab CI/CD
When parsing scripts for gitlab CI/CD files be aware that every element
in a list sequence gets treated as single script.

## Scriptcheck Directive
In case you want to force running scriptcheck over a specific yaml node
you can use our custom directive:

```yaml
test:
  # scriptcheck
  some-script: |
    cd $EXAMPLE
```

You can also specify the underlying shell in order to pass this argument
when running shellcheck for this specific yaml node:

````yaml
job_example:
  # scriptcheck shell=sh
  script:
    cd $EXAMPLE
````
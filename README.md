> [!CAUTION]
> This repository has been in-lined (using git-subtree) into [winc-release](https://github.com/cloudfoundry/winc-release/pull/46). Please make any
> future contributions directly to winc-release.

# Certificate Injector

### importing

This repository should be imported as `code.cloudfoundry.org/cert-injector`.

### testing

```
ginkgo -r -p -race .
```

### dependencies

```
dep ensure
```

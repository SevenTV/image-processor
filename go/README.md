# Go

## 1. To build this repo, you must first install all required dependancies

Update your apt repo

```
sudo apt-get update
```

Install all these packages used for building.

```
sudo apt-get install \
    ca-certificates \
    build-essential
```

## 2. Build the application

Now that everything else is installed you can simply run this to build the application.

```
make
```

The output folder should be created in `./out`

## 3. Clean up

If you wish to clean up you can run

```
make clean
```

## 4. Formatting

You need to install the dev deps

```
make dev_deps
```

and then you can do

```
make lint
```

## 5. Testing

You can run all the unit tests by doing

```
make test
```

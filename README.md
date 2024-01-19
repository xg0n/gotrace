## gotrace
### tracing for go programs

gotrace annotates function calls in go source files with log statements on entry and exit.

	usage: gotrace [flags] [path ...]
	-enableByDefault
		enable loging when the program start (default true)
	-exclude string
		exclude any matching functions, takes precedence over filter
	-exported
		only annotate exported functions
	-filter string
		only annotate functions matching the regular expression (default ".")
	-formatLength int
		limit the formatted length of each argumnet to 'size' (default 1024)
	-outputFile string
		file path to store output log or 'stdout', 'stderr' (default "stderr")
	-package
		show package name prefix on function calls
	-prefix string
		log prefix
	-returns
		show function return
	-timing
		print function durations. Implies -returns
	-toggleSignalNum int
		signal number to toggle enable/disable log write, default is 10 (SIGUSR1) (default 10)
	-w	re-write files in place

#### Example

    # gotrace operates directly on go source files.
    # Insert gotrace logging statements into all *.go files in the current directory
	# Make sure all files are saved in version control, as this rewrites them in-place!

    $ gotrace -outputFile /tmp/gotrace.log -returns -w -package *.go

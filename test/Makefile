TESTDIRS = filecontent filename package evrtest

all:

runtests:
	for x in $(TESTDIRS); do \
		$(MAKE) -C $$x runtests || exit 1; \
	done

clean:
	for x in $(TESTDIRS); do \
		$(MAKE) -C $$x clean || exit 1; \
	done

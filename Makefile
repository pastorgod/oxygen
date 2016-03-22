
all:
	sh ./install

.PHONY: table
table:
	sh ./install TABLE

.PHONY: cmd
cmd:
	sh ./install CMD

clean:
	sh ./clean

.PHONY: release
release:
	# create folder
	mkdir -p ./release/release
	rm -r ./release/release
	mkdir ./release/release
	# copy server
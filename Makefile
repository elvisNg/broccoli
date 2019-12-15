

.PHONY:  ALL tools

ALL:


tools: gen_broccoli


gen_broccoli:
	GOOS=linux go build -o tools/bin/ ./tools/gen-broccoli
	GOOS=windows go build -o tools/bin/ ./tools/gen-broccoli

errdef:
	gen-broccoli -onlybroccolierr -errdef errors/errdef.proto -dest .
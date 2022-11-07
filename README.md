# `erltf`

> Create and interpret information encoded with the [Erlang External Term Format][erlang-ext-tf].

This module is intended for communication specifically with the [Discord Gateway][discord-gateway],
usage outside this purpose is not within the scope of this library.

*A general rule of thumb is that types that require a Node identifier are not implemented.*

> **Note**
> Full functionality in accordance with the [Erlang External Term Format][erlang-ext-tf] is not off
> the table, and we will accept contributions that provide this functionality *without hindering
> **erltf**'s ability to communicate with* the [Discord Gateway][discord-gateway].

---

> In the event we do support the entire format, we will likely re-brand this module as a
> general-purpose [Erlang External Term Format][erlang-ext-tf] library. It is only currently
> advertised as having a specific purpose as to not lead those looking for a general implementation
> astray.

## Table of Contents

- [`erltf`](#erltf)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
    - [Encoding](#encoding)
    - [Decoding](#decoding)

## Features

### Encoding

- [X] `bool` values
- [X] untyped `nil`, interface-typed `nil` & `nil` pointer values
- [ ] values of types that implement the `EncodeETF` interface.
    > **Note**
    > This feature is waiting on [Decoding](#decoding) to be in full swing before implementing, so
    > that returned data can be verified.

- [X] pointers and interfaces (treated as recursion, does not serializes pointers as they are)
- [X] `uint`, `uint8`, `uint16` &`uint64` values (`uintptr` will likely never be supported)
- [X] `int`, `int8`, `int16` & `int64` values
- [X] `string` values
- [X] `[...]T` (array) & `[]T` (slice) values
- [X] `map[string]T` values
- [X] composite (struct) values
  - Not using an `erltf` tag on a field will cause **erltf** to use the field's name.
  - Supplying a tag of `erltf:"tag_name"` or `erltf:"-"` will rename or ignore the field
    respectively.

- [X] Explicitly encode `[]byte` data as `BINARY_EXT` with `Encoder.EncodeAsBinaryETF`.

### Decoding

- [ ] `bool` values
- [ ] values of types that implement the `DecodeETF` interface.
- [ ] pointers (treated as recursion, does not deserialize pointers as they are)
- [ ] `uint`, `uint8`, `uint16` &`uint64` values (`uintptr` will likely never be supported)
- [ ] `int`, `int8`, `int16` & `int64` values
- [ ] `string` values
- [ ] `[...]T` (array) & `[]T` (slice) values
- [ ] `map[string]T` values
- [ ] composite (struct) values
  - Not using an `erltf` tag on a field will cause **erltf** to use the field's name.
  - Supplying a tag of `erltf:"tag_name"` or `erltf:"-"` will rename or ignore the field
    respectively.

---

> 👋 While not required, we'd greatly **appreciate kudos in projects that use this module**. Have
> fun out there!

[erlang-ext-tf]: https://www.erlang.org/doc/apps/erts/erl_ext_dist.html
[discord-gateway]: https://discord.com/developers/docs/topics/gateway#gateway

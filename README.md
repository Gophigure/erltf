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
  - [Planned](#planned)

## Features

### Encoding

- `bool` values.
- untyped `nil`, interface-typed `nil` & `nil` pointer values.
- Non-`nil` pointer values.
- `uint?` (excluding `uintptr`) & `int?` values.
    <br>
    > **Note**
    > The `?` means `uint`, `int` and all related types (e.g. `uint8` or `int64`) are supported.

- `string`, `[...]T` & `[]T` values.
- `map[string]T` & struct values.
  - **erltf** supports field tags and behaves similar to the `encoding/json` package, it does not
    currently impose any naming schemes.
    - No tag means **erltf** will use the field's name.
    - `erltf:"encoded_pair_name"` will use `encoded_pair_name`.
    - `erltf:"-"` will cause **erltf** to ignore the field.

- Ability to force the encoding of `[]byte` values as `BINARY_EXT` via `Encoder.EncodeAsBinaryETF`.
  - You can force the encoding of all `string` values to `BINARY_EXT` with
    `AlwaysEncodeStringsToBinary`.

## Planned

- Allow declaring custom encode functions that override default behavior for custom types, like
  `encoding/json.Marshaler`.
- Decoding ETF data with regards to the below.
- Full functionality **in regards to** communicating with the [Discord Gateway][discord-gateway].

[erlang-ext-tf]: https://www.erlang.org/doc/apps/erts/erl_ext_dist.html
[discord-gateway]: https://discord.com/developers/docs/topics/gateway#gateway

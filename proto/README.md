# protocol generator

This used to use Google protobufs, but I switched to my own thing for fun.

`smash.d.ts` is used as a schema that decribes the protocol.  Note that it
isn't the actual types used in either TS or Go, but rather we reuse TS for
the parser and some semantic verification.

The wire format is currently undocumented while I figure out the details, but
it's a pretty straightforward serialization.

## Design goals

- Restrict the input format to make serialization/deserialization easy.  We don't
  need to support the full type system of TypeScript.
- No versioning-related negotiation; we can assume the client and server always
  were built from exactly the same version of the code.
- Generate native-feeling code in each language, even if that means the per-language
  generated APIs don't match each other.

## Design notes

### TypeScript

I wanted to make enumerated types ("Foo is either A or B") into plain unions in
TypeScript:

```
type Foo = A | B;
```

But this ends up failing because when you want to send such a mesage, you
want to mark which arm you chose, so the message-sending function needs some
runtime representation of `Foo` as distinct from `A` and `B`.
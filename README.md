
## Development

### Profiling

```sh
cargo build --release
perf record --call-graph=dwarf ./target/release/smash
perf report -g --comm smash
```

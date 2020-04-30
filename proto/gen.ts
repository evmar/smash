import * as ts from 'typescript';
import * as fs from 'fs';

class UnhandledError extends Error {
  constructor(readonly diag: ts.Diagnostic) {
    super('unhandled');
  }
}

function unhandled(node: ts.Node): never {
  const start = node.getStart();
  const end = node.getEnd();
  const diag: ts.Diagnostic = {
    start: node.pos,
    length: end - start,
    file: undefined, // will be filled in by 'catch'er
    code: 0,
    category: ts.DiagnosticCategory.Error,
    messageText: `unhandled ${ts.SyntaxKind[node.kind]}`,
  };
  throw new UnhandledError(diag);
}

function cap(name: string) {
  return name.substring(0, 1).toUpperCase() + name.substring(1);
}

function genGo(decls: proto.Named[], write: (out: string) => void) {
  const unions = new Set<string>();

  write(`package proto

import (
  "fmt"
  "io"
  "bufio"
)

type Msg interface{
  Write(w io.Writer) error
  Read(r *bufio.Reader) error
}

func ReadBoolean(r *bufio.Reader) (bool, error) {
  val, err := ReadUint8(r)
  if err != nil { return false, err }
  return val == 1, nil
}

func ReadUint8(r *bufio.Reader) (byte, error) {
  return r.ReadByte()
}

func ReadUint16(r *bufio.Reader) (uint16, error) {
  b1, err := r.ReadByte()
  if err != nil { return 0, err }
  b2, err := r.ReadByte()
  if err != nil { return 0, err }
  return (uint16(b1)<<8)|uint16(b2), nil
}

func ReadString(r *bufio.Reader) (string, error) {
  n, err := ReadUint16(r)
  if err != nil { return "", err }
  buf := make([]byte, n)
  _, err = io.ReadFull(r, buf)
  if err != nil { return "", err }
  return string(buf), nil
}

func WriteBool(w io.Writer, val bool) error {
  if val {
    return WriteUint8(w, 1)
  } else {
    return WriteUint8(w, 0)
  }
}
func WriteUint8(w io.Writer, val byte) error {
  buf := [1]byte{val}
  _, err := w.Write(buf[:])
  return err
}
func WriteUint16(w io.Writer, val uint16) error {
  buf := [2]byte{byte(val >> 8), byte(val & 0xFF)}
  _, err := w.Write(buf[:])
  return err
}
func WriteString(w io.Writer, str string) error {
  if len(str) >= 1<<16 {
    panic("overlong")
  }
  if err := WriteUint16(w, uint16(len(str))); err != nil { return err }
  _, err := w.Write([]byte(str))
  return err
}
`);
  for (const { name, type } of decls) {
    switch (type.kind) {
      case 'union':
        write(`type ${name} struct {\n`);
        write(`// ${type.types.map((t) => t.type).join(', ')}\n`);
        write(`Alt Msg\n`);
        write(`}\n`);
        unions.add(name);
        break;
      case 'struct':
        write(`type ${name} struct {\n`);
        for (const f of type.fields) {
          write(`${cap(f.name)} ${typeToGo(f.type)}\n`);
        }
        write(`}\n`);
        break;
      default:
        throw new Error(`todo: ${JSON.stringify(type)}`);
    }
  }

  for (const { name, type } of decls) {
    write(`func (msg *${name}) Write(w io.Writer) error {\n`);
    switch (type.kind) {
      case 'union':
        write(`switch alt := msg.Alt.(type) {\n`);
        type.types.forEach((t, i) => {
          write(`case *${t.type}:\n`);
          write(
            `if err := WriteUint8(w, ${i + 1}); err != nil { return err }\n`
          );
          write(`return alt.Write(w)\n`);
        });
        write(`}\n`);
        write(`panic("notimpl")\n`);
        break;
      case 'struct':
        for (const f of type.fields) {
          writeValue(f.type, `msg.${cap(f.name)}`);
        }
        write(`return nil\n`);
        break;
      default:
        throw new Error(`todo: ${JSON.stringify(type)}`);
    }
    write(`}\n`);
  }

  for (const { name, type } of decls) {
    write(`func (msg *${name}) Read(r *bufio.Reader) error {\n`);
    switch (type.kind) {
      case 'union':
        write(`alt, err := r.ReadByte()\n`);
        write(`if err != nil { return err }\n`);
        write(`switch alt {\n`);
        type.types.forEach((t, i) => {
          write(`case ${i + 1}:\n`);
          write(`var val ${t.type}\n`);
          write(`if err := val.Read(r); err != nil { return err }\n`);
          write(`msg.Alt = &val\n`);
          write(`return nil\n`);
        });
        write(
          `default: return fmt.Errorf("bad tag %d when reading ${name}", alt)`
        );
        write(`}\n`);
        break;
      case 'struct':
        write(`var err error\n`);
        write(`err = err\n`);
        for (const field of type.fields) {
          readValue(field.type, `msg.${cap(field.name)}`);
        }
          write(`return nil\n`);
          break;
      default:
        write(`panic("notimpl")\n`);
    }
    write(`}\n`);
  }

  function readValue(type: proto.Type, name: string) {
    let fn = `${name}.Read`;
    switch (type.kind) {
      case 'ref':
        switch (type.type) {
          case 'boolean':
          case 'uint8':
          case 'uint16':
          case 'string':
            write(`${name}, err = Read${cap(type.type)}(r)\n`);
            write(`if err != nil { return err }\n`)
            return;
        }
        break;
      case 'array':
        write(`{\n`);
        write(`n, err := ReadUint16(r)\n`);
        write(`if err != nil { return err }\n`);
        write(`var val ${typeToGo(type.type)}\n`)
        write(`for i := 0; i < int(n); i++ {\n`);
        readValue(type.type, 'val');
        write(`${name} = append(${name}, val)\n`)
        write(`}\n`);
        write(`}\n`);
        return;
    }
    write(`if err := ${fn}(r); err != nil { return err }\n`)
  }

  function writeValue(type: proto.Type, name: string) {
    switch (type.kind) {
      case 'ref':
        let fn = `Write${cap(type.type)}`;
        switch (type.type) {
          case 'boolean':
            fn = 'WriteBool';
          case 'uint8':
          case 'uint16':
          case 'string':
            write(`if err := ${fn}(w, ${name}); err != nil { return err }\n`);
            return;
        }
        break;
      case 'array':
        write(
          `if err := WriteUint8(w, uint8(len(${name}))); err != nil { return err }\n`
        );
        write(`for _, val := range ${name} {\n`);
        writeValue(type.type, 'val');
        write(`}\n`);
        return;
    }
    write(`if err := ${name}.Write(w); err != nil { return err }\n`);
  }

  function refToGo(ref: proto.Ref): string {
    switch (ref.type) {
      case 'boolean':
        return 'bool';
      default:
        return ref.type;
    }
  }
  function typeToGo(type: proto.Type): string {
    switch (type.kind) {
      case 'ref':
        return refToGo(type);
      case 'array':
        return `[]${typeToGo(type.type)}`;
      default:
        throw new Error(`todo: ${JSON.stringify(type)}`);
    }
  }
}

function genTS(decls: proto.Named[], write: (out: string) => void) {
  write(`
export type uint8 = number;
export type uint16 = number;
`);
  for (const { name, type } of decls) {
    switch (type.kind) {
      case 'union':
        write(`export class ${name} {\n`);
        write(` constructor(public alt: ${typeStr(type)}) {}\n`);
        write(`}\n`);
        break;
      case 'struct':
        write(`export class ${name} {\n`);
        for (const { name, type: fType } of type.fields) {
          write(`${name}!: ${typeStr(fType)};\n`);
        }
        write(
          `constructor(fields: ${name}) { Object.assign(this, fields); }\n`
        );
        write(`}\n`);
        break;
      default:
        throw new Error(`todo: ${JSON.stringify(type)}`);
    }
  }

  write(`export class Reader {
private ofs = 0;
constructor(readonly view: DataView) {}

private readUint8(): number {
  return this.view.getUint8(this.ofs++);
}

private readUint16(): number {
  const val = this.view.getUint16(this.ofs);
  this.ofs += 2;
  return val;
}

private readBoolean(): boolean {
  return this.readUint8() !== 0;
}

private readBytes(): DataView {
  const len = this.readUint16();
  const slice = new DataView(this.view.buffer, this.ofs, len);
  this.ofs += len;
  return slice;
}

private readString(): string {
  const bytes = this.readBytes();
  return new TextDecoder().decode(bytes);
}

private readArray<T>(elem: () => T): T[] {
  const len = this.readUint8();
  const arr: T[] = [];
  for (let i = 0; i < len; i++) {
    arr.push(elem());
  }
  return arr;
}
`);
  for (const { name, type } of decls) {
    write(`read${name}(): ${name} {\n`);
    switch (type.kind) {
      case 'union':
        write(`switch (this.readUint8()) {\n`);
        type.types.forEach((t, i) => {
          write(`case ${i + 1}: return new ${name}(${readRef(t)});\n`);
        });
        write(`default: throw new Error('parse error');`);
        write(`}\n`);
        break;
      case 'struct':
        write(`return new ${name}({\n`);
        for (const field of type.fields) {
          write(`${field.name}: `);
          switch (field.type.kind) {
            case 'ref':
              write(readRef(field.type));
              break;
            case 'array':
              write(`this.readArray(() => ${readRef(field.type.type)})`);
              break;
            default:
              throw new Error(`todo: ${JSON.stringify(field.type)}`);
          }
          write(`,\n`);
        }
        write(`});\n`);
        break;
      default:
        throw new Error(`todo: ${JSON.stringify(type)}`);
    }
    write(`}\n`);
  }
  write(`}\n`);

  write(`export class Writer {
    public ofs = 0;
    public buf = new Uint8Array();
    writeBoolean(val: boolean) {
      this.writeUint8(val ? 1 : 0);
    }
    writeUint8(val: number) {
      if (val > 0xFF) throw new Error('overflow');
      this.buf[this.ofs++] = val;
    }
    writeUint16(val: number) {
      if (val > 0xFFFF) throw new Error('overflow');
      this.buf[this.ofs++] = (val & 0xFF00) >> 8;
      this.buf[this.ofs++] = val & 0xFF;
    }
    writeString(str: string) {
      this.writeUint16(str.length);
      for (let i = 0; i < str.length; i++) {
        this.buf[this.ofs++] = str.charCodeAt(i);
      }
    }
    writeArray<T>(arr: T[], f: (t: T) => void) {
      this.writeUint16(arr.length);
      for (const elem of arr) {
        f(elem);
      }
    }
  `);
  for (const { name, type } of decls) {
    write(`write${name}(msg: ${name}) {\n`);
    switch (type.kind) {
      case 'union':
        type.types.forEach((t, i) => {
          if (i > 0) write(`else `);
          write(`if (msg.alt instanceof ${t.type}) {\n`);
          i;
          write(`this.writeUint8(${i + 1});\n`);
          write(`this.write${t.type}(msg.alt);\n`);
          write(`}\n`);
        });
        write(`else { throw new Error('unhandled case'); }\n`);
        break;
      case 'struct':
        for (const field of type.fields) {
          switch (field.type.kind) {
            case 'ref':
              write(`${writeRef(field.type)}(msg.${field.name});\n`);
              break;
            case 'array':
              write(
                `this.writeArray(msg.${field.name}, (val) => { ${writeRef(
                  field.type.type
                )}(val); });\n`
              );
              break;
            default:
              console.error(field);
              throw new Error('f');
          }
        }
        break;
      default:
        throw new Error(`todo: ${JSON.stringify(type)}`);
    }
    write(`}\n`);
  }
  write(`}\n`);

  function readRef(type: proto.Ref): string {
    return `this.read${cap(type.type)}()`;
  }

  function writeRef(type: proto.Ref): string {
    return `this.write${cap(type.type)}`;
  }

  function refStr(type: proto.Ref): string {
    switch (type.type) {
      case 'uint8':
      case 'uint16':
        return 'number';
      case 'boolean':
      case 'string':
      default:
        return type.type;
    }
  }
  function typeStr(type: proto.Type): string {
    // Intentionally not recursive -- we don't want complex nested types.
    switch (type.kind) {
      case 'ref':
        return refStr(type);
      case 'array':
        return `${refStr(type.type)}[]`;
      case 'union':
        return `${type.types.map(refStr).join('|')}`;
      default:
        throw new Error(`disallowed: ${JSON.stringify(type)}`);
    }
  }
}

namespace proto {
  export type Type = Ref | Array | Union | Struct;
  export interface Named {
    name: string;
    type: Type;
  }
  export interface Ref {
    kind: 'ref';
    type: string;
  }
  export interface Array {
    kind: 'array';
    type: Ref;
  }
  export interface Union {
    kind: 'union';
    types: Ref[];
  }
  export interface Struct {
    kind: 'struct';
    fields: Named[];
  }
}

function parse(sourceFile: ts.SourceFile): proto.Named[] {
  const decls: proto.Named[] = [];
  for (const stmt of sourceFile.statements) {
    if (ts.isTypeAliasDeclaration(stmt)) {
      const name = stmt.name.text;
      const type = parseType(stmt.type);
      if (type.kind === 'ref') continue;
      decls.push({ name, type });
    } else if (ts.isInterfaceDeclaration(stmt)) {
      const name = stmt.name.text;
      const fields = stmt.members.map((field) => {
        if (!field.name) unhandled(stmt);
        if (!ts.isIdentifier(field.name)) unhandled(field.name);
        const name = field.name.text;
        if (!ts.isPropertySignature(field)) unhandled(field);
        if (!field.type) unhandled(field);
        return { name, type: parseType(field.type) };
      });
      decls.push({ name, type: { kind: 'struct', fields } });
    } else {
      unhandled(stmt);
    }
  }
  return decls;

  function parseType(type: ts.TypeNode): proto.Type {
    if (type.kind === ts.SyntaxKind.NumberKeyword) {
      return { kind: 'ref', type: 'number' };
    } else if (type.kind === ts.SyntaxKind.StringKeyword) {
      return { kind: 'ref', type: 'string' };
    } else if (type.kind === ts.SyntaxKind.BooleanKeyword) {
      return { kind: 'ref', type: 'boolean' };
    } else if (ts.isTypeReferenceNode(type)) {
      if (!ts.isIdentifier(type.typeName)) unhandled(type.typeName);
      return { kind: 'ref', type: type.typeName.text };
    } else if (ts.isArrayTypeNode(type)) {
      const inner = parseType(type.elementType);
      if (inner.kind !== 'ref') unhandled(type.elementType);
      return { kind: 'array', type: inner };
    } else if (ts.isUnionTypeNode(type)) {
      const union = type.types.map((t) => {
        const inner = parseType(t);
        if (inner.kind !== 'ref') unhandled(t);
        return inner;
      });
      return { kind: 'union', types: union };
    } else {
      unhandled(type);
    }
  }
}

function usage() {
  console.error('usage: gen ts|go input.d.ts > output');
}

function main(args: string[]): number {
  const [mode, fileName] = args;

  if (!fileName) {
    usage();
    return 1;
  }

  const sourceText = fs.readFileSync(fileName, 'utf8');
  const sourceFile = ts.createSourceFile(
    fileName,
    sourceText,
    ts.ScriptTarget.Latest,
    true
  );

  let decls: proto.Named[];
  try {
    decls = parse(sourceFile);
  } catch (err) {
    if (err instanceof UnhandledError) {
      err.diag.file = sourceFile;
      const host: ts.FormatDiagnosticsHost = {
        getCurrentDirectory: () => process.cwd(),
        getCanonicalFileName: (fileName: string) => fileName,
        getNewLine: () => '\n',
      };
      console.error(ts.formatDiagnosticsWithColorAndContext([err.diag], host));
      return 1;
    }
    throw err;
  }

  if (mode === 'ts') {
    genTS(decls, (text) => process.stdout.write(text));
  } else if (mode == 'go') {
    genGo(decls, (text) => process.stdout.write(text));
  } else {
    usage();
    return 1;
  }

  return 0;
}

process.exitCode = main(process.argv.slice(2));

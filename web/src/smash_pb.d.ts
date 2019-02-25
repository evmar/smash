// package: proto
// file: smash.proto

import * as jspb from 'google-protobuf';

export class RunRequest extends jspb.Message {
  getCell(): number;
  setCell(value: number): void;

  getCommand(): string;
  setCommand(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RunRequest.AsObject;
  static toObject(
    includeInstance: boolean,
    msg: RunRequest
  ): RunRequest.AsObject;
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> };
  static extensionsBinary: {
    [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>;
  };
  static serializeBinaryToWriter(
    message: RunRequest,
    writer: jspb.BinaryWriter
  ): void;
  static deserializeBinary(bytes: Uint8Array): RunRequest;
  static deserializeBinaryFromReader(
    message: RunRequest,
    reader: jspb.BinaryReader
  ): RunRequest;
}

export namespace RunRequest {
  export type AsObject = {
    cell: number;
    command: string;
  };
}

export class Output extends jspb.Message {
  getCell(): number;
  setCell(value: number): void;

  hasText(): boolean;
  clearText(): void;
  getText(): string;
  setText(value: string): void;

  hasExitCode(): boolean;
  clearExitCode(): void;
  getExitCode(): number;
  setExitCode(value: number): void;

  getOutputCase(): Output.OutputCase;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Output.AsObject;
  static toObject(includeInstance: boolean, msg: Output): Output.AsObject;
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> };
  static extensionsBinary: {
    [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>;
  };
  static serializeBinaryToWriter(
    message: Output,
    writer: jspb.BinaryWriter
  ): void;
  static deserializeBinary(bytes: Uint8Array): Output;
  static deserializeBinaryFromReader(
    message: Output,
    reader: jspb.BinaryReader
  ): Output;
}

export namespace Output {
  export type AsObject = {
    cell: number;
    text: string;
    exitCode: number;
  };

  export enum OutputCase {
    OUTPUT_NOT_SET = 0,
    TEXT = 2,
    EXIT_CODE = 3
  }
}

export class ServerMsg extends jspb.Message {
  hasOutput(): boolean;
  clearOutput(): void;
  getOutput(): Output | undefined;
  setOutput(value?: Output): void;

  getMsgCase(): ServerMsg.MsgCase;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServerMsg.AsObject;
  static toObject(includeInstance: boolean, msg: ServerMsg): ServerMsg.AsObject;
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> };
  static extensionsBinary: {
    [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>;
  };
  static serializeBinaryToWriter(
    message: ServerMsg,
    writer: jspb.BinaryWriter
  ): void;
  static deserializeBinary(bytes: Uint8Array): ServerMsg;
  static deserializeBinaryFromReader(
    message: ServerMsg,
    reader: jspb.BinaryReader
  ): ServerMsg;
}

export namespace ServerMsg {
  export type AsObject = {
    output?: Output.AsObject;
  };

  export enum MsgCase {
    MSG_NOT_SET = 0,
    OUTPUT = 1
  }
}

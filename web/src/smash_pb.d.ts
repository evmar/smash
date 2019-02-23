// package: proto
// file: smash.proto

import * as jspb from 'google-protobuf';

export class RunRequest extends jspb.Message {
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
    command: string;
  };
}

export class OutputResponse extends jspb.Message {
  getText(): string;
  setText(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OutputResponse.AsObject;
  static toObject(
    includeInstance: boolean,
    msg: OutputResponse
  ): OutputResponse.AsObject;
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> };
  static extensionsBinary: {
    [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>;
  };
  static serializeBinaryToWriter(
    message: OutputResponse,
    writer: jspb.BinaryWriter
  ): void;
  static deserializeBinary(bytes: Uint8Array): OutputResponse;
  static deserializeBinaryFromReader(
    message: OutputResponse,
    reader: jspb.BinaryReader
  ): OutputResponse;
}

export namespace OutputResponse {
  export type AsObject = {
    text: string;
  };
}

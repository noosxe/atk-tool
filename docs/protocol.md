# Protocol Specifications: ATK Peripheral Tool

This document outlines the low-level USB HID communication protocol used to query and interface with ATK gaming peripherals.

---

## 🛰️ Transport Layer

Communication with ATK peripherals is performed over USB raw Human Interface Device (HID) interfaces using **Feature Reports**.

- **Report ID**: Packets sent to the device are prepended with a `1-byte` Report ID (typically `0x08`).
- **Payload Frame Size**: The underlying data payload is exactly `16 bytes`.
- **Total Write Packet Size**: `17 bytes` (`1 byte` Report ID + `16 bytes` Payload).
- **Read Buffer Size**: The response read buffer from the device is `64 bytes`.

---

## 📝 Write Packet Frame Layout

The 17-byte transmission packet sent to the device is organized as follows:

| Byte Index | Field Name | Description | Value / Details |
|---|---|---|---|
| **0** | Report ID | ID designating the HID feature report | `0x08` (Default) |
| **1** | Opcode | Subcommand identifying the target operation | `0x04` ([CmdQueryBattery](file:///home/mechsoull/Projects/atk-tool/protocol.go#L11)) |
| **2 - 15** | Payload Body | Command-specific parameter bytes | Padded with `0x00` |
| **16** | Checksum | CRC/Validation byte | Computed validation sum |

---

## 🧮 Checksum Algorithm

The checksum is the final byte (index 15 of the payload frame / index 16 of the overall packet). It is calculated so that the **8-bit wrapping sum of all 16 payload bytes equals `0x4D`**.

### Formula:
$$\text{Checksum} = 0\text{x}4\text{D} - \sum_{i=0}^{14} \text{Payload}[i]$$

### Example (Battery Query):
- Opcode: `0x04`
- Body bytes (14 bytes): `0x00, 0x00, ..., 0x00`
- Sum of first 15 bytes = `0x04`
- Checksum = `0x4D - 0x04 = 0x49`
- Resulting 16-byte payload:
  `[0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x49]`

---

## 📥 Read Response Frame Layout

After a write command is sent, the device replies with a 64-byte frame. For telemetry queries, the device must return at least **10 bytes**. The layout is parsed as follows:

```
+---------------+---------------+-------------------+-----------------------+
| Bytes 0 - 5   | Byte 6        | Byte 7            | Bytes 8 - 9           |
+---------------+---------------+-------------------+-----------------------+
| Header/Status | Battery %     | Padding / Reserved| Battery Voltage (mV)  |
| (6 bytes)     | (0 - 100)     | (1 byte)          | (Big-Endian uint16)   |
+---------------+---------------+-------------------+-----------------------+
```

### Parsing Logic details:

#### 1. Battery Percentage
- Location: Byte index `6` (`inBuf[6]`).
- Type: `uint8`.
- Value: `0` to `100` representing charge percentage.

#### 2. Battery Voltage
- Location: Byte indices `8` and `9` (`inBuf[8]` and `inBuf[9]`).
- Type: 16-bit Big-Endian Unsigned Integer.
- Unit: Millivolts (mV).
- **Decoding Formula**:
  $$\text{Voltage (Volts)} = \frac{(\text{inBuf}[8] \ll 8) \mid \text{inBuf}[9]}{1000.0}$$

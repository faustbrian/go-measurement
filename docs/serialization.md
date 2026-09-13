# Serialization And Adapters

Quantity JSON is an object with decimal text and unit metadata:

```json
{"value":"12.50","unit":"kg"}
```

The decimal is always a string to prevent JSON-number precision loss. XML uses
`<quantity><value>12.50</value><unit>kg</unit></quantity>`. `driver.Valuer` and
`sql.Scanner` store the same JSON document in SQL text or JSON columns.
`Dimensions` JSON and XML preserve each side's original unit plus quantity.

Direct JSON decoding rejects numbers, unknown fields, trailing values,
duplicate fields, unsupported units, invalid decimals, and documents beyond
`MaxSerializedBytes`. Text uses canonical symbols and `MaxTextBytes`.

`ParseQuantityXML` and `ParseDimensionsXML` enforce `MaxSerializedBytes`, the
fixed schema depth and field counts, a token ceiling, and scalar-field limits.
They reject attributes, namespaces, comments, directives, processing
instructions, CDATA, references, duplicate or unknown fields, nested scalar
content, and trailing documents. Errors do not echo attacker-controlled XML.

`Quantity.UnmarshalXML` and `Dimensions.UnmarshalXML` now fail closed with
`ErrUnboundedXML`. `encoding/xml` parses a start token before invoking those
callbacks, so they cannot safely establish a pre-token byte limit. Migrate raw
`xml.Unmarshal` calls to the package parse functions or `adapters/wire`.

The optional `adapters/wire` package uses the selected wire codec's configured
byte limit and routes XML through the core bounded parser. It supports only
explicit JSON and XML formats. Apply a transport limit before constructing the
input byte slice. The former `measurementwire` path delegates to this
implementation and remains available for compatibility.

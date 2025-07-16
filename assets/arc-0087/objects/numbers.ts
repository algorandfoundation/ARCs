export function parseBigInt(valueString: string) {
    let value: string | number;
    value = parseInt(valueString);
    if (isNaN(value)) {
        value = valueString;
    }
    return value;
}

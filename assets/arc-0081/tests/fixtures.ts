import { PartKeyIntegrityHashArgs } from "../src";

export const emptyBase64length32 = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
export const emptyBase64length64 =
  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==";
export const zeroAddress =
  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAY5HFKQ";

export const emptyFixture: PartKeyIntegrityHashArgs = {
    genesisHashB64: emptyBase64length32,
    account: zeroAddress,
    selectionKeyB64: emptyBase64length32,
    voteKeyB64: emptyBase64length32,
    stateProofKeyB64: emptyBase64length64,
    voteFirstValid: 0,
    voteLastValid: 0,
    keyDilution: 0,
}
export const emptyFixtureIntegrityHash = "J5UPK3SPRM2PY";

export const fixture1: PartKeyIntegrityHashArgs = {
  genesisHashB64: "kUt08LxeVAAGHnh4JoAoAMM9ql/hBwSoiFtlnKNeOxA=",
  account: "ROBOTMMVHPOETOTAX3J26UXYKVZX6QB7FHHYGBC44JNBUXMTABD5I3CODE",
  selectionKeyB64: "AniTlT3G6bQ5svv4HcmIp82TfmabTH7QbUxwTnFAgEU=",
  voteKeyB64: "nmnpvyyj9W2jtjzo41zHEGwnY1R03v1wwwlLuqWVYJk=",
  stateProofKeyB64:
    "NHS/AFnNW9UvaZeO+d84Dt8IaL84QXKAjZ9Fd0Pp+Jhh8eWEaIhWVGVdkw4pmJ5+/+zHgNSx5Dhxzv+LRFO6rw==",
  voteFirstValid: 10000,
  voteLastValid: 20000,
  keyDilution: 100,
}
export const fixture1IntegrityHash = "S4626H5HM6F7Y";
export const fixture1ZeroFirstValidIntegrityHash = "NDNRRGSZREUXI";

export const fixture2: PartKeyIntegrityHashArgs = {
  genesisHashB64: "kUt08LxeVAAGHnh4JoAoAMM9ql/hBwSoiFtlnKNeOxA=",
  account: "OHQTAISSIGRGIGVN6TVJ6WYLBHFTUC437T4E2LRRXGWVNJ4GSZOXKPH7N4",
  selectionKeyB64: "e4kBLu7zXOorjLVzJHOiAn+IhOBsYBCqqHKaJCiCdJs=",
  voteKeyB64: "WWHePYtNZ2T3sHkqdd/38EvoFWrnIKPrTo6xN/4T1l4=",
  stateProofKeyB64:
    "1GdNPOck+t6yXvuXxrDEPKqgi4I2sTaNugV1kd5ksUW2G1U6x1FT0WR3aT3ZSSmbYoDt3cVrB3vIPJA8GkqSYg==",
  voteFirstValid: 3118965,
  voteLastValid: 4104516,
  keyDilution: 993,
}
export const fixture2IntegrityHash = "2P4RW3CRB7VKA";
export const fixture2OfflineIntegrityHash = "W5TM4AVRNRJIC";

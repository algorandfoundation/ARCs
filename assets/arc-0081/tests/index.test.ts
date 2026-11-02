import { describe, expect, test } from "@jest/globals";
import {
  getPartKeyIntegrityHash,
  getOfflinePartKeyIntegrityHash,
  PartKeyIntegrityHashArgs,
} from "../src/index";
import onlineFixtures from "./online.fixtures.json";
import offlineFixtures from "./offline.fixtures.json";

describe("partkey integrity hash", () => {
  for (const [fixtureName, { params: fixture, expected }] of Object.entries(
    onlineFixtures
  )) {
    test(`online partkey integrity hash should be correctly calculated (${fixtureName})`, () => {
      expect(getPartKeyIntegrityHash(fixture)).toBe(expected);
    });
  }

  for (const [fixtureName, { params: fixture, expected }] of Object.entries(
    offlineFixtures
  )) {
    test(`offline partkey integrity hash should be correctly calculated (${fixtureName})`, () => {
      expect(getOfflinePartKeyIntegrityHash(fixture)).toBe(expected);
    });
  }

  test("partkey integrity hash should work with unpadded base64 strings", () => {
    const { params: fixtureBase, expected } = onlineFixtures.fixture1;
    const stripRegex = /=*$/;
    const fixture: PartKeyIntegrityHashArgs = {
      genesisHashB64: fixtureBase.genesisHashB64.replace(stripRegex, ""),
      account: fixtureBase.account,
      selectionKeyB64: fixtureBase.selectionKeyB64.replace(stripRegex, ""),
      voteKeyB64: fixtureBase.voteKeyB64.replace(stripRegex, ""),
      stateProofKeyB64: fixtureBase.stateProofKeyB64.replace(stripRegex, ""),
      voteFirstValid: fixtureBase.voteFirstValid,
      voteLastValid: fixtureBase.voteLastValid,
      keyDilution: fixtureBase.keyDilution,
    };
    expect(getPartKeyIntegrityHash(fixture)).toBe(expected);
  });

  test("offline partkey integrity hash should match empty online partkey integrity hash", () => {
    const { params: emptyFixtureOnline, expected } = onlineFixtures.empty;
    const { params: emptyFixtureOffline } = offlineFixtures.empty;

    const one = getPartKeyIntegrityHash(emptyFixtureOnline);
    const two = getOfflinePartKeyIntegrityHash(emptyFixtureOffline);

    expect(one).toBe(two);
    expect(one).toBe(expected);
  });

  test("it should work without voteFirstValid", () => {
    const { params: fixture, expected } = onlineFixtures.fixture1ZeroFirstValid;
    const withFirst = getPartKeyIntegrityHash(fixture);
    // @ts-ignore
    delete fixture.voteFirstValid;
    const withoutFirst = getPartKeyIntegrityHash(fixture);

    expect(withFirst).toBe(expected);
    expect(withoutFirst).toBe(expected);
  });

  for (const field of [
    "genesisHashB64",
    "account",
    "selectionKeyB64",
    "voteKeyB64",
    "stateProofKeyB64",
    "voteLastValid",
    "keyDilution",
  ]) {
    test(`it should fail without ${field}`, () => {
      const { params: fixtureBase } = onlineFixtures.fixture1;

      const fixture: PartKeyIntegrityHashArgs = {
        ...fixtureBase,
      };
      // @ts-ignore
      delete fixture[field];
      expect(() => getPartKeyIntegrityHash(fixture)).toThrow(field);
    });
  }

  const lengthCheckfields: Array<keyof PartKeyIntegrityHashArgs> = [
    "genesisHashB64",
    "account",
    "selectionKeyB64",
    "voteKeyB64",
    "stateProofKeyB64",
  ];
  for (const field of lengthCheckfields) {
    test(`it should fail with invalid ${field}`, () => {
      const { params: fixtureBase } = onlineFixtures.fixture1;

      const fixture: PartKeyIntegrityHashArgs = {
        ...fixtureBase,
      };
      // @ts-ignore
      fixture[field] = fixture[field].slice(2);
      expect(() => getPartKeyIntegrityHash(fixture)).toThrow(field);
    });
  }
});

import test from "node:test";import assert from "node:assert/strict";
import {defaults,validate,timecode,clipLabel} from "../frontend/js/project.js";
test("default project validates",()=>assert.deepEqual(validate(defaults()),defaults()));
test("invalid format rejected",()=>assert.throws(()=>validate({...defaults(),format:"bad"})));
test("invalid scale rejected",()=>assert.throws(()=>validate({...defaults(),scale:1000})));
test("timecode",()=>assert.equal(timecode(125.8),"02:05"));
test("clip label strips markup",()=>assert.equal(clipLabel("<a>"),"a"));

test("projects saved before trimming use the complete clip", () => {
 const {trimStart, trimEnd, ...legacy} = defaults();
 assert.deepEqual(validate(legacy), defaults());
});
test("project trim bounds are validated and round-trip through JSON", () => {
 const trimmed = {...defaults(), trimStart: 1.1, trimEnd: 1.7};
 assert.deepEqual(validate(JSON.parse(JSON.stringify(trimmed))), trimmed);
 assert.throws(() => validate({...trimmed, trimEnd: 1}));
 assert.throws(() => validate({...trimmed, trimEnd: Infinity}));
});

test("splitting, skipping a gap and legacy trim produce the same ordered ranges", async () => {
 const {sourceClips,splitClips,nextClip} = await import("../frontend/js/project.js");
 const original = sourceClips({...defaults(),trimStart:1,trimEnd:4},10);
 assert.deepEqual(splitClips(original,2),{index:1,clips:[{start:1,end:2},{start:2,end:4}]});
 assert.equal(splitClips(original,4),null);
 const gapped = [{start:1,end:2},{start:3,end:4}];
 assert.deepEqual(nextClip(gapped,2.5),gapped[1]);
 assert.equal(nextClip(gapped,4),null);
 assert.throws(()=>validate({...defaults(),clips:[{start:1,end:3},{start:2,end:4}]}));
});

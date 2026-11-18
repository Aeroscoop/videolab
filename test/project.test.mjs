import test from "node:test";import assert from "node:assert/strict";
import {defaults,validate,timecode,clipLabel} from "../frontend/js/project.js";
test("default project validates",()=>assert.deepEqual(validate(defaults()),defaults()));
test("invalid format rejected",()=>assert.throws(()=>validate({...defaults(),format:"bad"})));
test("invalid scale rejected",()=>assert.throws(()=>validate({...defaults(),scale:1000})));
test("timecode",()=>assert.equal(timecode(125.8),"02:05"));
test("clip label strips markup",()=>assert.equal(clipLabel("<a>"),"a"));

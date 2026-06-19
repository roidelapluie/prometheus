// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { describe, expect, it } from "vitest";
import { docsVersionPath } from "./docsURL";

describe("docsVersionPath", () => {
  it("keeps the major.minor of a release version", () => {
    expect(docsVersionPath("3.12.0")).toBe("3.12");
    expect(docsVersionPath("2.55.1")).toBe("2.55");
  });

  it("keeps the major.minor of a pre-release version", () => {
    expect(docsVersionPath("3.13.0-rc.0")).toBe("3.13");
  });

  it("falls back to latest for an empty version", () => {
    expect(docsVersionPath("")).toBe("latest");
  });

  it("falls back to latest for an unreplaced placeholder", () => {
    expect(docsVersionPath("VERSION_PLACEHOLDER")).toBe("latest");
  });
});

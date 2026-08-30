import Foundation

public struct LoomOwnedRegionUpdate: Codable, Sendable {
  public var path: String
  public var regionID: String
  public var changed: Bool
}

public struct LoomOwnedRegionUpdater: Sendable {
  public init() {}

  public func replaceXAMLRegion(
    path: String,
    regionID: String,
    content: String,
    createIfMissing: Bool = false
  ) throws
    -> LoomOwnedRegionUpdate
  {
    let url = URL(fileURLWithPath: path).standardizedFileURL
    try validateRegionID(regionID)
    guard let existing = try? String(contentsOf: url, encoding: .utf8) else {
      guard createIfMissing else {
        throw LoomError.unreadableSource(path)
      }
      let initialized = initializedXAMLRegion(regionID: regionID, content: content)
      try FileManager.default.createDirectory(
        at: url.deletingLastPathComponent(),
        withIntermediateDirectories: true
      )
      try initialized.write(to: url, atomically: true, encoding: .utf8)
      return LoomOwnedRegionUpdate(path: url.path, regionID: regionID, changed: true)
    }
    guard !createIfMissing || containsRegionMarkers(existing, regionID: regionID) else {
      throw LoomError.invalidArguments(
        "--init-region only creates missing files; existing files still require explicit markers.")
    }
    let updated = try replaceXAMLRegion(in: existing, regionID: regionID, content: content)
    if updated != existing {
      try updated.write(to: url, atomically: true, encoding: .utf8)
    }
    return LoomOwnedRegionUpdate(path: url.path, regionID: regionID, changed: updated != existing)
  }

  public func initializedXAMLRegion(regionID: String, content: String) -> String {
    let id = regionID.trimmingCharacters(in: .whitespacesAndNewlines)
    return """
      <Grid xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"
            xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml">
        <!-- LOOM-BEGIN \(id) -->
      \(trimTrailingNewline(content))
        <!-- LOOM-END \(id) -->
      </Grid>

      """
  }

  public func containsRegionMarkers(_ existing: String, regionID: String) -> Bool {
    let id = regionID.trimmingCharacters(in: .whitespacesAndNewlines)
    guard isSafeRegionID(id) else { return false }
    return existing.contains("<!-- LOOM-BEGIN \(id) -->")
      || existing.contains("<!-- LOOM-END \(id) -->")
  }

  public func replaceXAMLRegion(in existing: String, regionID: String, content: String) throws
    -> String
  {
    let id = regionID.trimmingCharacters(in: .whitespacesAndNewlines)
    guard isSafeRegionID(id) else {
      throw LoomError.invalidArguments("Region id must use letters, numbers, dot, underscore, or hyphen.")
    }

    let begin = "<!-- LOOM-BEGIN \(id) -->"
    let end = "<!-- LOOM-END \(id) -->"
    let beginRanges = ranges(of: begin, in: existing)
    let endRanges = ranges(of: end, in: existing)
    guard beginRanges.count == 1 && endRanges.count == 1 else {
      throw LoomError.invalidArguments(
        "Expected exactly one LOOM-BEGIN and LOOM-END marker for region \(id).")
    }
    guard let beginRange = beginRanges.first, let endRange = endRanges.first,
      beginRange.upperBound <= endRange.lowerBound
    else {
      throw LoomError.invalidArguments("Owned region \(id) has malformed marker order.")
    }

    let replacement = "\n\(trimTrailingNewline(content))\n"
    return String(existing[..<beginRange.upperBound])
      + replacement
      + String(existing[endRange.lowerBound...])
  }

  private func ranges(of needle: String, in haystack: String) -> [Range<String.Index>] {
    var result: [Range<String.Index>] = []
    var cursor = haystack.startIndex
    while let range = haystack.range(of: needle, range: cursor..<haystack.endIndex) {
      result.append(range)
      cursor = range.upperBound
    }
    return result
  }

  private func isSafeRegionID(_ value: String) -> Bool {
    !value.isEmpty
      && value.allSatisfy { $0.isLetter || $0.isNumber || $0 == "." || $0 == "_" || $0 == "-" }
  }

  private func validateRegionID(_ value: String) throws {
    let id = value.trimmingCharacters(in: .whitespacesAndNewlines)
    guard isSafeRegionID(id) else {
      throw LoomError.invalidArguments("Region id must use letters, numbers, dot, underscore, or hyphen.")
    }
  }

  private func trimTrailingNewline(_ value: String) -> String {
    var result = value
    while result.hasSuffix("\n") || result.hasSuffix("\r") {
      result.removeLast()
    }
    return result
  }
}

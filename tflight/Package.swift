// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "TFlight",
    platforms: [.macOS(.v14)],
    products: [.executable(name: "TFlight", targets: ["TFlight"])],
    targets: [.executableTarget(name: "TFlight")]
)

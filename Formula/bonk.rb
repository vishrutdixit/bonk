class Bonk < Formula
  desc "LLM-powered spaced repetition CLI for technical interview prep"
  homepage "https://github.com/vishrutdixit/bonk"
  version "0.2.3"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/vishrutdixit/bonk/releases/download/v0.2.3/bonk-darwin-arm64.tar.gz"
      sha256 "e03bc265010ecd8b695f140c7722eff7bb62f86091fc1dd64645b844dc5638c7"
    end
    on_intel do
      url "https://github.com/vishrutdixit/bonk/releases/download/v0.2.3/bonk-darwin-amd64.tar.gz"
      sha256 "bbd312e935df819aad309253f2494d2d0768a8e12b8eb2c70db932e33b3628bc"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/vishrutdixit/bonk/releases/download/v0.2.3/bonk-linux-arm64.tar.gz"
      sha256 "6f317e68a12eb17f4b9105b7ecb103b2acf876d6644231802d932ab1672a783e"
    end
    on_intel do
      url "https://github.com/vishrutdixit/bonk/releases/download/v0.2.3/bonk-linux-amd64.tar.gz"
      sha256 "394c6f90258d34fd4f2ee2c1f693851ff0b37f174c5a890b5ff994a00440c485"
    end
  end

  # Voice mode dependencies (optional but recommended on macOS)
  depends_on "sox" => :recommended
  depends_on "whisper-cpp" => :recommended

  def install
    # Binary name in archive includes platform suffix
    binary_name = "bonk-#{OS.kernel_name.downcase}-#{Hardware::CPU.arch == :arm64 ? "arm64" : "amd64"}"
    bin.install binary_name => "bonk"
  end

  def post_install
    # Create bonk directory for data
    (var/"bonk").mkpath

    # Download whisper model for voice mode (macOS only)
    if OS.mac? && build.with?("whisper-cpp")
      bonk_dir = Pathname.new(Dir.home)/".bonk"
      bonk_dir.mkpath
      model_path = bonk_dir/"ggml-tiny.en.bin"
      unless model_path.exist?
        ohai "Downloading whisper model for voice mode..."
        system "curl", "-sSL",
          "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin",
          "-o", model_path
      end
    end
  end

  def caveats
    <<~EOS
      To use bonk, you need an Anthropic API key:
        export ANTHROPIC_API_KEY="your-key-here"

      Start drilling:
        bonk

      Voice mode (macOS):
        bonk --voice
        - Press space to record your answer
        - Press 's' to skip text-to-speech

      If voice mode isn't working, run:
        bonk setup
    EOS
  end

  test do
    assert_match "bonk", shell_output("#{bin}/bonk --version")
  end
end

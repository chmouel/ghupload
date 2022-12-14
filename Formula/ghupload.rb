# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Ghupload < Formula
  desc "ghupload  - smee.io go client"
  homepage "https://github.com/chmouel/ghupload"
  version "0.1.1"

  on_macos do
    url "https://github.com/chmouel/ghupload/releases/download/v0.1.1/ghupload_0.1.1_MacOS_all.tar.gz"
    sha256 "993d4137a97ac76f0d8e294b266009fb3e00e9dc3aa70c062a8be916bd9a79fd"

    def install
      bin.install "ghupload" => "ghupload"
      output = Utils.popen_read("SHELL=bash #{bin}/ghupload completion bash")
      (bash_completion/"ghupload").write output
      output = Utils.popen_read("SHELL=zsh #{bin}/ghupload completion zsh")
      (zsh_completion/"_ghupload").write output
      prefix.install_metafiles
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/chmouel/ghupload/releases/download/v0.1.1/ghupload_0.1.1_Linux_arm64.tar.gz"
      sha256 "5e83a2333ba596fded92082c129863187562616dee7f1bd08bbff63df038168e"

      def install
        bin.install "ghupload" => "ghupload"
        output = Utils.popen_read("SHELL=bash #{bin}/ghupload completion bash")
        (bash_completion/"ghupload").write output
        output = Utils.popen_read("SHELL=zsh #{bin}/ghupload completion zsh")
        (zsh_completion/"_ghupload").write output
        prefix.install_metafiles
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/chmouel/ghupload/releases/download/v0.1.1/ghupload_0.1.1_Linux_x86_64.tar.gz"
      sha256 "d8f9af213016edb0b87d0ccba4a163b8e86db3b98543e6980eadaafe258541a0"

      def install
        bin.install "ghupload" => "ghupload"
        output = Utils.popen_read("SHELL=bash #{bin}/ghupload completion bash")
        (bash_completion/"ghupload").write output
        output = Utils.popen_read("SHELL=zsh #{bin}/ghupload completion zsh")
        (zsh_completion/"_ghupload").write output
        prefix.install_metafiles
      end
    end
  end
end

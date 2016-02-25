class Goshe < FPM::Cookery::Recipe
  name 'goshe'

  version '0.3'
  revision '1'
  description 'Small utility to send stats to Datadog.'

  homepage 'https://github.com/darron/goshe'
  source "https://github.com/darron/#{name}/releases/download/v#{version}/#{name}-#{version}-linux-amd64.zip"
  sha256 '6e808b4a54b2239c64f62d4e3faaed852425aaf0dc646f8bf6636e4b1bf56c20'

  maintainer 'Darron <darron@froese.org>'
  vendor 'octohost'

  license 'Mozilla Public License, version 2.0'

  conflicts 'goshe'
  replaces 'goshe'

  build_depends 'unzip'

  def build
    safesystem "mkdir -p #{builddir}/usr/local/bin/"
    safesystem "cp -f #{builddir}/#{name}-#{version}-linux-amd64/#{name}-#{version}-linux-amd64 #{builddir}/usr/local/bin/#{name}"
  end

  def install
    safesystem "mkdir -p #{destdir}/usr/local/bin/"
    safesystem "cp -f #{builddir}/usr/local/bin/#{name} #{destdir}/usr/local/bin/#{name}"
    safesystem "chmod 755 #{destdir}/usr/local/bin/#{name}"
  end
end

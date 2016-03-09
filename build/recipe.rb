class Goshe < FPM::Cookery::Recipe
  name 'goshe'

  version '0.4'
  revision '1'
  description 'Small utility to send stats to Datadog.'

  homepage 'https://github.com/darron/goshe'
  source "https://github.com/darron/#{name}/releases/download/v#{version}/#{name}-#{version}-linux-x86_64.zip"
  sha256 '983a1a87c5ff37675963c02af8ad05c0027c26d2453c8e3d7a58da012ef240f2'

  maintainer 'Darron <darron@froese.org>'
  vendor 'octohost'

  license 'Mozilla Public License, version 2.0'

  conflicts 'goshe'
  replaces 'goshe'

  build_depends 'unzip'

  def build
    safesystem "mkdir -p #{builddir}/usr/local/bin/"
    safesystem "cp -f #{builddir}/#{name}-#{version}-linux-x86_64/#{name}-#{version}-linux-x86_64 #{builddir}/usr/local/bin/#{name}"
  end

  def install
    safesystem "mkdir -p #{destdir}/usr/local/bin/"
    safesystem "cp -f #{builddir}/usr/local/bin/#{name} #{destdir}/usr/local/bin/#{name}"
    safesystem "chmod 755 #{destdir}/usr/local/bin/#{name}"
  end
end

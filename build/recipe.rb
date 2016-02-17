class Goshe < FPM::Cookery::Recipe
  name 'goshe'

  version '0.2'
  revision '1'
  description 'Small utility to send stats to Datadog.'

  homepage 'https://github.com/darron/goshe'
  source "https://github.com/darron/#{name}/releases/download/v#{version}/#{name}-#{version}-linux-amd64.zip"
  sha256 'f316c7511f4525f83055ee92b2924318d39e023cfd47a13eb70e530d08879766'

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

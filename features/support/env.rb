require 'bundler/setup'

require 'fileutils'

require 'aruba/cucumber'
require 'wrong'

World(Wrong)
World(FileUtils)

Dir['features/fixtures/_env/*'].each do |var|
  ENV[File.basename(var)] = File.read(var).strip
end

Before do
  fixtures_dir = File.realpath('features/fixtures')
  in_current_dir do
    ln_s fixtures_dir, 'fixtures'
  end
end


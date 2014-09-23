require 'bundler/setup'

require 'fileutils'

require 'aruba/cucumber'
require 'wrong'

World(Wrong)
World(FileUtils)

Before do
  fixtures_dir = File.realpath('features/fixtures')
  in_current_dir do
    ln_s fixtures_dir, 'fixtures'
  end
end


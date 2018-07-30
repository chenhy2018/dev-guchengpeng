gem 'capistrano-scm-jenkins', '0.0.8'
gem 'capistrano-subset-fix'

require 'capistrano-scm-jenkins'
require 'capistrano/ext/multistage'
require 'capistrano-subset-fix'

#set :default_stage, :sandbox

set :application, "base"

set :scm, :jenkins
set :jenkins_use_netrc, true

set :use_sudo, false


# use portal.test.qiniu.com to visit

set :local_user, `whoami`.strip()

namespace :deploy do
  task :more_update, :roles => :app do
    run "chmod a+x #{release_path}/qbox*"
    run "echo '#{local_user} deployed #{application} to #{stage}' > #{release_path}/DEPLOY-NOTES"
    run "test -f #{shared_path}/config/#{conf_name}.conf && ln -sfn #{shared_path}/config/#{conf_name}.conf #{release_path}/#{conf_name}.conf"
    run "ln -sfn #{shared_path}/run #{release_path}/run"
  end
  task :more_setup do
    run "mkdir -p #{shared_path}/config"
    run "mkdir -p #{shared_path}/run"
    run "mkdir -p #{shared_path}/run/auditlog"
  end
end

before "deploy:finalize_update", "deploy:more_update"
after "deploy:setup", "deploy:more_setup"
